FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bot

FROM alpine:latest

WORKDIR /app

RUN apt-get update && apt-get install -y tzdata

COPY --from=builder /app/bot .

EXPOSE 8080

CMD ["./bot"]
