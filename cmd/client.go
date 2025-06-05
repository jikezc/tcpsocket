package main

import (
	"fmt"
	"tcpsocketv2/intranal/socket"
)

func main() {
	client := socket.NewClient("127.0.0.1", 8080)
	if err := client.Connect(); err != nil {
		fmt.Printf("Connect error: %v", err)
		return
	}
	client.Run()
}
