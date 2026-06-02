FROM golang:1.26-alpine3.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o server ./cmd/api

FROM alpine:3.22
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
RUN addgroup -S app && adduser -S app -G app && chown app:app /app/server
USER app
EXPOSE 3003
CMD ["./server"]
