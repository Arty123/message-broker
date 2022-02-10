FROM golang:1.16.13-alpine
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o ./bin/message_broker ./cmd/message-broker/main.go
ENTRYPOINT ["/app/bin/message_broker", "8080"]