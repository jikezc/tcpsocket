package handler

import (
	"fmt"
	"net"
	"tcpsocketv2/internal/serializer"
	"tcpsocketv2/internal/socket"
	message "tcpsocketv2/pb"
	"tcpsocketv2/pkg/utils"
)

// HeartbeatReq 发送心跳包
func (h *ClientMsgHandler) HeartbeatReq() error {
	os := utils.GetOS()
	cpu, men := utils.GetPerformance()
	// 创建心跳包
	heartbeat := &message.MSG_HEARTBEAT{
		Os:  &os,
		Cpu: &cpu,
		Mem: &men,
	}
	// 序列化心跳包
	pkg, err := serializer.SerializeMessage(
		message.CommandType_CommandType_Heartbeat,
		heartbeat,
	)
	if err != nil {
		fmt.Println("序列化心跳包异常: ", err)
		return fmt.Errorf("序列化心跳包异常, Error: %v\n", err)
	}
	n, err := h.Client.Conn.Write(pkg)
	if err != nil {
		return fmt.Errorf("发送心跳包异常: %v\n", err)
	}
	fmt.Printf("发送心跳包成功，发送了%v字节数据\n", n)
	return nil
}

// HandleHeartbeatReq 处理心跳包
func (h *ServerMsgHandler) HandleHeartbeatReq(conn net.Conn, payload *message.MSG_HEARTBEAT) error {
	_session, ok := h.Server.GetSession(conn)
	if !ok {
		return fmt.Errorf("未找到对应的会话")
	}

	// 修改会话信息
	_session.ClientSpec = socket.Spec{
		Os:  *payload.Os,
		Cpu: *payload.Cpu,
		Mem: *payload.Mem,
	}
	_session.LastAliveTime = utils.GetCurrentTimestamp()

	// 更新会话
	h.Server.UpdateSession(conn, _session)

	return nil
}
