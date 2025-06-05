FROM golang:1.22

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o app .

ENV PORT=8080
CMD ["/app/app"]