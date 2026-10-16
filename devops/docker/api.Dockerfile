FROM golang:1.24-alpine AS builder

WORKDIR /src/be/services/gateway
COPY be/platform /src/be/platform
COPY be/services/gateway/go.mod be/services/gateway/go.sum ./
RUN go mod download
COPY be/services/gateway/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/gateway

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S -G app app
COPY --from=builder /out/api /usr/local/bin/api
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
