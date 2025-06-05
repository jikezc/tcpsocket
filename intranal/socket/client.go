package socket

import (
	"bufio"
	"fmt"
	"net"
	"tcpsocketv2/intranal/handler"
	"tcpsocketv2/intranal/serializer"
	message "tcpsocketv2/pb"
	"tcpsocketv2/pkg/enums"
)

// ClientMsgHandler 客户端消息处理器
type ClientMsgHandler interface {
	HandshakeReq(conn net.Conn) error
	HeartbeatReq(conn net.Conn, payload *message.MSG_HEARTBEAT) error
}

// Client 客户端
type Client struct {
	Address string
	Port    int
	Status  enums.ClientStatusEM
	Conn    net.Conn
	handler ClientMsgHandler
}

// NewClient 创建客户端
func NewClient(address string, port int) *Client {
	return &Client{
		Address: address,
		Port:    port,
		Status:  enums.ClientStatusWaiting,
		handler: nil,
	}
}

// Connect 连接到服务器
func (c *Client) Connect() error {
	server := fmt.Sprintf("%s:%d", c.Address, c.Port)
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return fmt.Errorf("Connect to %v Failed\nerr: %v\n", server, err)
	}
	c.Conn = conn
	fmt.Printf("Connect to %v Success\n", server)
	return nil
}

func (c *Client) RegisterHandler(handler ClientMsgHandler) {
	c.handler = handler
}

// Run 启动客户端
func (c *Client) Run() {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			fmt.Printf("Conn Close Error: %v\n", err)
		}
	}()
	// 发送握手消息
	err := c.handler.HandshakeReq(c.Conn)
	if err != nil {
		fmt.Println("客户端发送握手消息失败了，Error: ", err)
		return
	}
	reader := bufio.NewReader(c.Conn)
	for {
		// 反序列化消息
		command, payload, err := serializer.DeserializeMessage(reader)
		if err != nil {
			fmt.Println("deserialize message error: ", err)
			break
		}
		fmt.Printf("收到服务器的响应，handler: %v, payload: %v\n", command, payload)
		switch command {
		case message.CommandType_CommandType_HandShakeResp:
			err := handler.HandshakeCallback(*c, payload.(*message.MSG_HANDSHAKE_RESP))
			if err != nil {
				fmt.Println("handshake callback error: ", err)
			}
		}
	}
}
