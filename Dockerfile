FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o marketflow ./cmd/marketflow/main.go