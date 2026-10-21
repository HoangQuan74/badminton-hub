FROM golang:1.24-alpine AS builder
WORKDIR /src/be/services/socket
COPY be/platform /src/be/platform
COPY contracts /src/contracts
COPY be/services/socket/go.mod be/services/socket/go.sum ./
RUN go mod download
COPY be/services/socket/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/socket ./cmd/socket

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
COPY --from=builder /out/socket /usr/local/bin/socket
USER app
EXPOSE 9093
ENTRYPOINT ["/usr/local/bin/socket"]
