# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
RUN apk add --no-cache postgresql16-client
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ./bin/server ./cmd/server

# Run stage
FROM alpine:3.21
RUN apk add --no-cache postgresql16-client ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/server .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
