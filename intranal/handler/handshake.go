package handler

import (
	"fmt"
	"net"
	"tcpsocketv2/intranal/serializer"
	"tcpsocketv2/intranal/socket"
	message "tcpsocketv2/pb"
	"tcpsocketv2/pkg/enums"
	"tcpsocketv2/pkg/utils"
)

// HandshakeCallback 握手响应回调函数
func HandshakeCallback(c socket.Client, payload *message.MSG_HANDSHAKE_RESP) error {
	fmt.Printf("握手响应回调函数执行, 返回码: %v, 消息内容: %v\n", *payload.Code, *payload.Message)
	if *payload.Code == 0 {
		handshakeSuccess(c)
	} else {
		handshakeFail(c)
	}
	return nil
}

// handshakeSuccess 握手成功
func handshakeSuccess(c socket.Client) {
	// 更新客户端状态
	c.Status = enums.ClientStatusConnected
	fmt.Printf("握手成功，开始发送心跳请求，当前客户端状态：%v\n", c.Status)
	StartHeartbeat(c)
}

// handshakeFail 握手失败
func handshakeFail(c socket.Client) {
	// 更新客户端状态
	c.Status = enums.ClientStatusDisConnected
	fmt.Printf("握手失败，客户端断开连接")
}

// HandshakeReq 发送握手请求
func (h *ClientMsgHandler) HandshakeReq(c socket.Client) {
	deviceId, err := utils.GetFQDN()
	fmt.Println("Get FQDN: ", deviceId)
	if err != nil {
		fmt.Println("Get FQDN Error: ", err)
		return
	}
	version := "1.0"
	// 序列化握手消息
	pkg, serializeErr := serializer.SerializeMessage(
		message.CommandType_CommandType_HandShakeReq,
		&message.MSG_HANDSHAKE_REQ{
			Version:  &version,
			DeviceId: &deviceId,
		},
	)
	if serializeErr != nil {
		fmt.Println("客户端序列化消息异常: ", err)
		return
	}

	n, err := c.Conn.Write(pkg)
	if err != nil {
		fmt.Printf("Send Error: %v\n", err)
		return
	}
	fmt.Printf("Send %v Bytes\n", n)
}

// HandlerHandshake 处理握手包
func (h *ServerMsgHandler) HandlerHandshake(conn net.Conn, payload *message.MSG_HANDSHAKE_REQ) error {
	// 将客户端加入SessionMap
	h.Server.SessionMap[conn] = socket.Session{
		DiverId:       payload.GetDeviceId(),
		LastAliveTime: utils.GetCurrentTimestamp(),
	}

	// 开始心跳检查
	h.Server.StartHeartbeatChecker(conn)

	code := int32(0)
	respMsg := "success"
	pkg, serializeErr := serializer.SerializeMessage(
		message.CommandType_CommandType_HandShakeResp,
		&message.MSG_HANDSHAKE_RESP{
			Code:    &code,
			Message: &respMsg,
		},
	)
	if serializeErr != nil {
		return fmt.Errorf("Server, 序列化消息异常: %v\n", serializeErr)
	}
	_, writeErr := conn.Write(pkg)
	if writeErr != nil {
		return fmt.Errorf("Server, 发送消息失败，关闭连接：%v\n", conn.RemoteAddr())
	}
	return nil
}
