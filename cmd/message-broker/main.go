package main

import (
	"github.com/message-broker/internal/command"
	"log"
)

func main() {
	err := command.CmdHttpServer.Execute()
	if err != nil {
		log.Fatal(err)
		return
	}
}
