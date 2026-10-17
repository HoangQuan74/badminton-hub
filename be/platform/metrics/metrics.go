// Package metrics provides small, dependency-free Prometheus metrics for
// process boundaries. It deliberately keeps labels bounded to HTTP method,
// registered route, and status code.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

const bucketCount = 11

var buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type requestKey struct {
	method string
	route  string
	status int
}

type durationKey struct {
	method string
	route  string
}

type histogram struct {
	counts [bucketCount]uint64
	count  uint64
	sum    uint64
}

// Registry stores metrics for one process. Create one registry per process so
// separate gateway instances cannot accidentally share state in tests.
type Registry struct {
	service   string
	inFlight  atomic.Int64
	mu        sync.Mutex
	requests  map[requestKey]uint64
	durations map[durationKey]*histogram
}

func New(service string) *Registry {
	return &Registry{service: service, requests: make(map[requestKey]uint64), durations: make(map[durationKey]*histogram)}
}

// Middleware records completed HTTP requests and an in-flight gauge.
func (r *Registry) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		r.inFlight.Add(1)
		defer func() {
			r.inFlight.Add(-1)
			status := c.Writer.Status()
			if status == 0 {
				status = http.StatusOK
			}
			route := c.FullPath()
			if route == "" {
				route = "__unknown__"
			}
			method := c.Request.Method
			r.mu.Lock()
			r.requests[requestKey{method: method, route: route, status: status}]++
			key := durationKey{method: method, route: route}
			h := r.durations[key]
			if h == nil {
				h = &histogram{}
				r.durations[key] = h
			}
			seconds := time.Since(started).Seconds()
			for i, boundary := range buckets {
				if seconds <= boundary {
					h.counts[i]++
				}
			}
			h.count++
			h.sum += uint64(seconds * 1e9)
			r.mu.Unlock()
		}()
		c.Next()
	}
}

// Handler exposes the registry in Prometheus text format.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = fmt.Fprintf(w, "# HELP http_requests_total Completed HTTP requests.\n# TYPE http_requests_total counter\n")
		r.mu.Lock()
		requests := make(map[requestKey]uint64, len(r.requests))
		for key, value := range r.requests {
			requests[key] = value
		}
		durations := make(map[durationKey]*histogram, len(r.durations))
		for key, value := range r.durations {
			copyValue := *value
			durations[key] = &copyValue
		}
		r.mu.Unlock()
		requestKeys := make([]requestKey, 0, len(requests))
		for key := range requests {
			requestKeys = append(requestKeys, key)
		}
		sort.Slice(requestKeys, func(i, j int) bool {
			return requestKeys[i].method+requestKeys[i].route < requestKeys[j].method+requestKeys[j].route
		})
		for _, key := range requestKeys {
			fmt.Fprintf(w, "http_requests_total{service=%q,method=%q,route=%q,status=%q} %d\n", r.service, key.method, key.route, strconv.Itoa(key.status), requests[key])
		}
		fmt.Fprintln(w, "# HELP http_request_duration_seconds HTTP request duration in seconds.")
		fmt.Fprintln(w, "# TYPE http_request_duration_seconds histogram")
		durationKeys := make([]durationKey, 0, len(durations))
		for key := range durations {
			durationKeys = append(durationKeys, key)
		}
		sort.Slice(durationKeys, func(i, j int) bool {
			return durationKeys[i].method+durationKeys[i].route < durationKeys[j].method+durationKeys[j].route
		})
		for _, key := range durationKeys {
			h := durations[key]
			for i, boundary := range buckets {
				fmt.Fprintf(w, "http_request_duration_seconds_bucket{service=%q,method=%q,route=%q,le=%q} %d\n", r.service, key.method, key.route, strconv.FormatFloat(boundary, 'f', -1, 64), h.counts[i])
			}
			fmt.Fprintf(w, "http_request_duration_seconds_bucket{service=%q,method=%q,route=%q,le=\"+Inf\"} %d\n", r.service, key.method, key.route, h.count)
			fmt.Fprintf(w, "http_request_duration_seconds_sum{service=%q,method=%q,route=%q} %.9f\n", r.service, key.method, key.route, float64(h.sum)/1e9)
			fmt.Fprintf(w, "http_request_duration_seconds_count{service=%q,method=%q,route=%q} %d\n", r.service, key.method, key.route, h.count)
		}
		fmt.Fprintln(w, "# HELP http_in_flight_requests Current in-flight HTTP requests.")
		fmt.Fprintln(w, "# TYPE http_in_flight_requests gauge")
		fmt.Fprintf(w, "http_in_flight_requests{service=%q} %d\n", r.service, r.inFlight.Load())
	})
}
