# stage: build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY pkg /app/pkg
COPY services/notification /app/services/notification

WORKDIR /app/services/notification

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /notification ./cmd/notification

# stage: runtime
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g "" appuser

WORKDIR /app

COPY --from=builder /notification /app/notification

USER appuser

ENTRYPOINT ["/app/notification"]
