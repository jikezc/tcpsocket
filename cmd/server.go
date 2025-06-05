package main

import (
	"fmt"
	"tcpsocketv2/intranal/handler"
	"tcpsocketv2/intranal/socket"
)

func main() {
	// 创建 Server 实例
	server := socket.NewServer("127.0.0.1", 8080)
	// 注册消息处理器
	server.RegisterHandler(handler.NewServerMsgHandler(server))

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("ListenAndServe error: %v", err)
		return
	}
}
