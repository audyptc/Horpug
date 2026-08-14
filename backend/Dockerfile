FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache tzdata ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/main ./main

USER appuser

EXPOSE 8080

CMD ["./main"]
