package handler

import (
	"tcpsocketv2/intranal/socket"
)

// ServerMsgHandler 服务端消息处理
type ServerMsgHandler struct {
	Server socket.Server
}

// NewServerMsgHandler 创建服务端消息处理
func NewServerMsgHandler(server socket.Server) *ServerMsgHandler {
	return &ServerMsgHandler{
		Server: server,
	}
}

// ClientMsgHandler 客户端消息处理
type ClientMsgHandler struct {
	Client socket.Client
}

// NewClientMsgHandler 创建客户端消息处理
func NewClientMsgHandler(client socket.Client) *ClientMsgHandler {
	return &ClientMsgHandler{
		Client: client,
	}
}
