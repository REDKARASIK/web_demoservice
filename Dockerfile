FROM golang:1.24.5 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/service

EXPOSE 8081

CMD ["./app"]
