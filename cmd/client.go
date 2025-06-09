package main

import (
	"fmt"
	"tcpsocketv2/common/logger"
	"tcpsocketv2/internal/handler"
	"tcpsocketv2/internal/socket"
)

func main() {
	client, cancel := socket.NewClient("127.0.0.1", 8080)
	defer cancel()
	l := logger.FromCtx(client.Ctx)
	if err := client.Connect(); err != nil {
		l.Error(fmt.Sprintf("Connect error: %v", err))
		return
	}
	client.RegisterHandler(handler.NewClientMsgHandler(client))
	client.Run()
}
