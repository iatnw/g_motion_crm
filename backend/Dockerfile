FROM golang:1.26-alpine

WORKDIR /app

RUN apk add --no-cache postgresql16-client

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

# Pre-build the binary so air has a fallback even if live recompile fails
COPY . .
RUN go build -o ./tmp/main ./cmd/server

CMD ["air"]
