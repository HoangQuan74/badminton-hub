package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	kafkaClient "github.com/segmentio/kafka-go"
)

// Config contains the connection settings shared by Kafka producers and consumers.
type Config struct {
	Brokers      []string
	ClientID     string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Client owns Kafka connection settings and the writers/readers created from them.
type Client struct {
	config  Config
	dialer  *kafkaClient.Dialer
	mu      sync.Mutex
	writers []*kafkaClient.Writer
	readers []*kafkaClient.Reader
}

// Open validates the broker list and checks that at least one broker is reachable.
func Open(ctx context.Context, config Config) (*Client, error) {
	config.Brokers = normalizeBrokers(config.Brokers)
	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("Kafka brokers are required")
	}
	if config.DialTimeout <= 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = 10 * time.Second
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = 10 * time.Second
	}

	dialer := &kafkaClient.Dialer{Timeout: config.DialTimeout, ClientID: config.ClientID}
	checkCtx, cancel := context.WithTimeout(ctx, config.DialTimeout)
	defer cancel()
	connection, err := dialer.DialContext(checkCtx, "tcp", config.Brokers[0])
	if err != nil {
		return nil, fmt.Errorf("dial Kafka broker: %w", err)
	}
	if err := connection.Close(); err != nil {
		return nil, fmt.Errorf("close Kafka broker probe: %w", err)
	}

	return &Client{config: config, dialer: dialer}, nil
}

// NewWriter creates a producer for one topic.
func (client *Client) NewWriter(topic string) (*kafkaClient.Writer, error) {
	if strings.TrimSpace(topic) == "" {
		return nil, fmt.Errorf("Kafka topic is required")
	}
	writer := kafkaClient.NewWriter(kafkaClient.WriterConfig{
		Brokers:      client.config.Brokers,
		Topic:        topic,
		Balancer:     &kafkaClient.Hash{},
		Dialer:       client.dialer,
		ReadTimeout:  client.config.ReadTimeout,
		WriteTimeout: client.config.WriteTimeout,
		RequiredAcks: int(kafkaClient.RequireOne),
	})
	client.mu.Lock()
	client.writers = append(client.writers, writer)
	client.mu.Unlock()
	return writer, nil
}

// NewReader creates a consumer for a topic and consumer group.
func (client *Client) NewReader(topic, groupID string) (*kafkaClient.Reader, error) {
	if strings.TrimSpace(topic) == "" {
		return nil, fmt.Errorf("Kafka topic is required")
	}
	if strings.TrimSpace(groupID) == "" {
		return nil, fmt.Errorf("Kafka consumer group is required")
	}
	reader := kafkaClient.NewReader(kafkaClient.ReaderConfig{
		Brokers:        client.config.Brokers,
		GroupID:        groupID,
		Topic:          topic,
		Dialer:         client.dialer,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	client.mu.Lock()
	client.readers = append(client.readers, reader)
	client.mu.Unlock()
	return reader, nil
}

// Close releases all readers and writers owned by the client.
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	var closeErr error
	for _, reader := range client.readers {
		if err := reader.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	for _, writer := range client.writers {
		if err := writer.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

func normalizeBrokers(brokers []string) []string {
	result := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		if value := strings.TrimSpace(broker); value != "" {
			result = append(result, value)
		}
	}
	return result
}
