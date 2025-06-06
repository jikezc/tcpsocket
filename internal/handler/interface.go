package handler

import (
	"tcpsocketv2/internal/socket"
)

// ServerMsgHandler 服务端消息处理
type ServerMsgHandler struct {
	Server socket.Server
}

// NewServerMsgHandler 创建服务端消息处理
func NewServerMsgHandler(server *socket.Server) *ServerMsgHandler {
	_handler := &ServerMsgHandler{
		Server: *server,
	}
	_handler.Server.Handler = _handler
	return _handler
}

// ClientMsgHandler 客户端消息处理
type ClientMsgHandler struct {
	Client socket.Client
}

// NewClientMsgHandler 创建客户端消息处理
func NewClientMsgHandler(client *socket.Client) *ClientMsgHandler {
	_handler := &ClientMsgHandler{
		Client: *client,
	}
	_handler.Client.Handler = _handler
	return _handler
}
