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
	
	l := logger.Get()
	// 创建 Server 实例
	server := socket.NewServer(cfg.SrvInfo.Host, cfg.SrvInfo.Port)
	// 注册消息处理器
	server.RegisterHandler(handler.NewServerMsgHandler(server))
	l.Info("Server started, register the handler of message successfully!")
	if err := server.ListenAndServe(); err != nil {
		l.Error(fmt.Sprintf("ListenAndServe error: %v", err))
		return
	}
}
