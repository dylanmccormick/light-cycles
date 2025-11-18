package main

import (
	"os"

	"github.com/dylanmccormick/light-cycles/client"
	"github.com/dylanmccormick/light-cycles/server"
)

func main() {
	firstArg := ""
	if len(os.Args) > 1 {
		firstArg = os.Args[1]
	}


	switch firstArg {
	case "run":
		server.Run()
	case "sim":
		client.Connect()
	case "tui":
		client.Tui()
	}
}
