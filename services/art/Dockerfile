# stage: build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY pkg /app/pkg
COPY proto /app/proto
COPY services/art /app/services/art

WORKDIR /app/services/art

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /art ./cmd/art

# stage: runtime
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g "" appuser

WORKDIR /app

COPY --from=builder /art /app/art

USER appuser

ENTRYPOINT ["/app/art"]
