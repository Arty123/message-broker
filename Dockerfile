FROM golang:1.14.3-alpine
RUN mkdir /app
ADD . /app
WORKDIR /apps
RUN go build -o ./bin/message_broker ./cmd/message-broker/main.go
ENTRYPOINT ["/app/bin/message_broker", "8080"]