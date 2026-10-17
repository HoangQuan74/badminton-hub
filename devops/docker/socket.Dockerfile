FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY be/platform/go.mod be/platform/go.sum be/platform/
COPY contracts/go.mod contracts/go.sum contracts/
COPY be/services/socket/go.mod be/services/socket/go.sum be/services/socket/
RUN cd be/services/socket && go mod download
COPY be/platform/ be/platform/
COPY contracts/ contracts/
COPY be/services/socket/ be/services/socket/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/socket ./be/services/socket/cmd/socket

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
COPY --from=builder /out/socket /usr/local/bin/socket
USER app
EXPOSE 9093
ENTRYPOINT ["/usr/local/bin/socket"]
