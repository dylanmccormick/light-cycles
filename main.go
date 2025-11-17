package main

import (
	"time"

	"github.com/dylanmccormick/light-cycles/client"
	"github.com/dylanmccormick/light-cycles/server"
)

func main() {
	go server.Run()
	time.Sleep(100 * time.Millisecond)
	go client.Connect()
	select {}
}
