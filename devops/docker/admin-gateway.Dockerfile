FROM golang:1.24-alpine AS builder

WORKDIR /src/be/gateways/admin
COPY be/platform /src/be/platform
COPY be/gateways/admin/go.mod be/gateways/admin/go.sum ./
RUN go mod download
COPY be/gateways/admin/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/admin-gateway ./cmd/admin-gateway

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S -G app app
COPY --from=builder /out/admin-gateway /usr/local/bin/admin-gateway
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/admin-gateway"]
