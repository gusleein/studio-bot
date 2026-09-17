FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -o /app/bin/studio-api \
    ./cmd/main.go

FROM alpine:3.20

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=builder /app/bin/studio-api ./studio-api
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/templates ./templates

RUN addgroup -g 1001 -S studio && \
    adduser -u 1001 -S studio -G studio && \
    chown -R studio:studio /app

USER studio

EXPOSE 8088

ENTRYPOINT ["./studio-api"]
