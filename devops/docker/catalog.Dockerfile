FROM golang:1.24-alpine AS builder

WORKDIR /src/be/services/catalog
COPY be/platform /src/be/platform
COPY contracts /src/contracts
COPY be/services/catalog/go.mod be/services/catalog/go.sum ./
RUN go mod download
COPY be/services/catalog/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/catalog ./cmd/catalog && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/catalog-bucket-init ./cmd/catalog-bucket-init

FROM alpine:3.22
RUN apk add --no-cache ca-certificates busybox-extras && addgroup -S app && adduser -S -G app app
COPY --from=builder /out/catalog /usr/local/bin/catalog
COPY --from=builder /out/catalog-bucket-init /usr/local/bin/catalog-bucket-init
USER app
EXPOSE 9094
ENTRYPOINT ["/usr/local/bin/catalog"]
