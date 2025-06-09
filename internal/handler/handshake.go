package handler

import (
	"context"
	"fmt"
	"net"
	"tcpsocketv2/common/enums"
	"tcpsocketv2/common/logger"
	"tcpsocketv2/internal/serializer"
	"tcpsocketv2/internal/socket"
	message "tcpsocketv2/pb"
	"tcpsocketv2/pkg/utils"
)

// handshakeSuccess 握手成功
func (h *ClientMsgHandler) handshakeSuccess() (err error) {
	// 更新客户端状态
	h.Client.Status = enums.ClientStatusConnected
	l := logger.FromCtx(h.Client.Ctx)
	l.Info(fmt.Sprintf("握手成功，开始发送心跳请求，当前客户端状态：%v", h.Client.Status))
	err = h.Client.StartHeartbeat()
	return err
}

// handshakeFail 握手失败
func (h *ClientMsgHandler) handshakeFail() (err error) {
	// 更新客户端状态
	h.Client.Status = enums.ClientStatusDisConnected
	return fmt.Errorf("握手失败，客户端断开连接")
}

// HandshakeReq 发送握手请求
func (h *ClientMsgHandler) HandshakeReq() error {
	l := logger.FromCtx(h.Client.Ctx)
	deviceId, err := utils.GetFQDN()
	l.Debug(fmt.Sprintf("Get FQDN: %v", deviceId))
	if err != nil {
		return fmt.Errorf("Get FQDN Error: %v\n", err)
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
		return fmt.Errorf("客户端序列化消息异常: %v\n", serializeErr)
	}
	n, writeErr := h.Client.Conn.Write(pkg)
	if writeErr != nil {
		return fmt.Errorf("Send Error: %v\n", err)
	}
	l.Debug(fmt.Sprintf("Send %v Bytes", n))
	return nil
}

func (h *ClientMsgHandler) HandleHandshakeResp(payload *message.MSG_HANDSHAKE_RESP) (err error) {
	l := logger.FromCtx(h.Client.Ctx)
	l.Debug(fmt.Sprintf("握手响应回调函数执行, 返回码: %v, 消息内容: %v", *payload.Code, *payload.Message))
	if *payload.Code == 0 {
		err = h.handshakeSuccess()
	} else {
		err = h.handshakeFail()
	}
	return err
}

// HandleHandshakeReq 处理握手包
func (h *ServerMsgHandler) HandleHandshakeReq(conn net.Conn, payload *message.MSG_HANDSHAKE_REQ, ctx context.Context) error {
	l := logger.FromCtx(ctx)
	// 将客户端加入SessionMap
	h.Server.SessionMap[conn] = socket.Session{
		DiverId:       payload.GetDeviceId(),
		LastAliveTime: utils.GetCurrentTimestamp(),
		Ctx:           ctx,
	}
	l.Debug(fmt.Sprintf("Server当前会话: %v", h.Server.SessionMap))
	// 开始心跳检查
	h.Server.StartHeartbeatChecker(conn)

	code := int32(0)
	respMsg := "success"
	respPayload := &message.MSG_HANDSHAKE_RESP{
		Code:    &code,
		Message: &respMsg,
	}
	pkg, serializeErr := serializer.SerializeMessage(
		message.CommandType_CommandType_HandShakeResp,
		respPayload,
	)
	if serializeErr != nil {
		return fmt.Errorf("Server, 序列化消息异常: %v\n", serializeErr)
	}
	_, writeErr := conn.Write(pkg)
	if writeErr != nil {
		return fmt.Errorf("Server, 发送消息失败，关闭连接：%v\n", conn.RemoteAddr())
	} else {
		l.Debug(fmt.Sprintf("Server, 发送握手响应消息成功，发送 %v Bytes, respPayload: %v", len(pkg), respPayload))
	}
	return nil
}
