FROM golang:1.24-alpine AS builder

WORKDIR /src/be/services/identity
COPY be/platform /src/be/platform
COPY contracts /src/contracts
COPY be/services/identity/go.mod be/services/identity/go.sum ./
RUN go mod download
COPY be/services/identity/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/identity ./cmd/identity

FROM alpine:3.22
RUN apk add --no-cache ca-certificates busybox-extras && addgroup -S app && adduser -S -G app app
COPY --from=builder /out/identity /usr/local/bin/identity
USER app
EXPOSE 9091
ENTRYPOINT ["/usr/local/bin/identity"]
