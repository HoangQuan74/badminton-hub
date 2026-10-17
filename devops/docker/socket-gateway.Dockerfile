FROM golang:1.24-alpine AS builder

WORKDIR /src/be/gateways/socket
COPY be/platform /src/be/platform
COPY contracts /src/contracts
COPY be/gateways/socket/go.mod be/gateways/socket/go.sum ./
RUN go mod download
COPY be/gateways/socket/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/socket-gateway ./cmd/socket-gateway

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S -G app app
COPY --from=builder /out/socket-gateway /usr/local/bin/socket-gateway
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/socket-gateway"]
