FROM golang:1.23

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o app .

ENV PORT=8080
CMD ["/app/app"]