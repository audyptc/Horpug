FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache tzdata ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
# Uploaded files live here; it is mounted as a volume in docker-compose.yml, and
# must be owned by appuser so the (non-root) app can write to it.
RUN mkdir -p /app/uploads && chown appuser:appgroup /app/uploads

WORKDIR /app

COPY --from=builder /app/main ./main

USER appuser

EXPOSE 8080

CMD ["./main"]
