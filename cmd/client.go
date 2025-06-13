package main

import (
	"fmt"
	"tcpsocketv2/common/logger"
	"tcpsocketv2/config"
	"tcpsocketv2/internal/handler"
	"tcpsocketv2/internal/socket"
)

func main() {
	// 初始化配置
	config.Init()
	// 获取配置
	cfg := config.Get()
	
	client, cancel := socket.NewClient(cfg.SrvInfo.Host, cfg.SrvInfo.Port)
	defer cancel()
	l := logger.FromCtx(client.Ctx)
	if err := client.Connect(); err != nil {
		l.Error(fmt.Sprintf("Connect error: %v", err))
		return
	}
	client.RegisterHandler(handler.NewClientMsgHandler(client))
	client.Run()
}
