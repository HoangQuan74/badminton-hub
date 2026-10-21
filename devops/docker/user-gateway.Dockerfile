FROM golang:1.24-alpine AS builder

WORKDIR /src/be/gateways/user
COPY be/platform /src/be/platform
COPY contracts /src/contracts
COPY be/gateways/user/go.mod be/gateways/user/go.sum ./
RUN go mod download
COPY be/gateways/user/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/user-gateway ./cmd/user-gateway

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S -G app app
COPY --from=builder /out/user-gateway /usr/local/bin/user-gateway
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/user-gateway"]
