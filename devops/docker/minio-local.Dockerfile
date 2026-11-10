# Local-only MinIO build. Upstream stopped distributing the community images.
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
RUN git clone https://github.com/minio/minio.git . && \
    git checkout 7aac2a2c5b7c882e68c1ce017d8256be2feea27f && \
    CGO_ENABLED=0 go build -trimpath -o /out/minio .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && addgroup -S minio && adduser -S -G minio minio && \
    mkdir -p /data && chown minio:minio /data
COPY --from=builder /out/minio /usr/local/bin/minio
USER minio
ENTRYPOINT ["/usr/local/bin/minio"]
CMD ["server", "/data"]
