# stage: build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY pkg /app/pkg
COPY proto /app/proto
COPY services/gateway /app/services/gateway

WORKDIR /app/services/gateway

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gateway ./cmd/gateway

# stage: runtime
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g "" appuser

WORKDIR /app

COPY --from=builder /gateway /app/gateway

USER appuser

ENTRYPOINT ["/app/gateway"]
