package handler

import (
	"fmt"
	"net"
	"tcpsocketv2/global"
	"tcpsocketv2/intranal/serializer"
	"tcpsocketv2/intranal/socket"
	message "tcpsocketv2/pb"
	"tcpsocketv2/pkg/enums"
	"tcpsocketv2/pkg/utils"
	"time"
)

// StartHeartbeat 启动心跳
func StartHeartbeat(c socket.Client) {
	// 启动心跳协程
	go func() {
		// 创建心跳定时器
		ticker := time.NewTicker(time.Second * global.HeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 检查连接状态
				if c.Status != enums.ClientStatusConnected {
					fmt.Println("客户端未连接，停止心跳发送")
					return
				}
				// 发送心跳包
				sendHeartbeat(c)
			}
		}
	}()
}

// sendHeartbeat 发送心跳包
func sendHeartbeat(c socket.Client) {
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
		return
	}
	n, err := c.Conn.Write(pkg)
	if err != nil {
		fmt.Println("发送心跳包异常: ", err)
	}
	fmt.Printf("发送心跳包成功，发送了%v字节数据\n", n)
}

// HandlerHeartbeat 处理心跳包
func (h *ServerMsgHandler) HandlerHeartbeat(conn net.Conn, payload *message.MSG_HEARTBEAT) error {
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
