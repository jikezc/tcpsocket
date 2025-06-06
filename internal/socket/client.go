package socket

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"tcpsocketv2/global"
	"tcpsocketv2/internal/serializer"
	message "tcpsocketv2/pb"
	"tcpsocketv2/pkg/enums"
	"time"
)

// ClientMsgHandlerInterface 客户端消息处理器
type ClientMsgHandlerInterface interface {
	HandshakeReq() error
	HandleHandshakeResp(payload *message.MSG_HANDSHAKE_RESP) error
	HeartbeatReq() error
}

// Client 客户端
type Client struct {
	Address string
	Port    int
	Status  enums.ClientStatusEM
	Conn    net.Conn
	Handler ClientMsgHandlerInterface
}

// NewClient 创建客户端
func NewClient(address string, port int) *Client {
	return &Client{
		Address: address,
		Port:    port,
		Status:  enums.ClientStatusWaiting,
		Conn:    nil,
		Handler: nil,
	}
}

// RegisterHandler 注册处理器
func (c *Client) RegisterHandler(handler ClientMsgHandlerInterface) {
	c.Handler = handler
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

// Run 启动客户端
func (c *Client) Run() {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			fmt.Printf("Conn Close Error: %v\n", err)
		}
	}()
	// 发送握手消息
	err := c.Handler.HandshakeReq()
	if err != nil {
		fmt.Println("客户端发送握手消息失败了，Error: ", err)
		return
	}
	reader := bufio.NewReader(c.Conn)

	for {
		// 反序列化消息
		command, payload, err := serializer.DeserializeMessage(reader)
		if err == io.EOF {
			// TODO: 后续实现关系连接挥手消息，以及断线重连
			fmt.Println("收到EOF，服务器关闭了连接，退出程序")
			break
		}
		if err != nil {
			fmt.Println("deserialize message error: ", err)
			continue
		}
		fmt.Printf("收到服务器的响应，handler: %v, payload: %v\n", command, payload)
		handleMsgErr := c.handleMessage(command, payload)
		if handleMsgErr != nil {
			fmt.Printf("处理服务器消息异常, Error: %v\n", handleMsgErr)
			continue
		}
	}
}

// handleMessage 处理消息
func (c *Client) handleMessage(command message.CommandType, payload interface{}) (err error) {
	switch command {
	case message.CommandType_CommandType_HandShakeResp:
		err = c.Handler.HandleHandshakeResp(payload.(*message.MSG_HANDSHAKE_RESP))
	default:
		err = fmt.Errorf("Unknow command: %v, payload: %v\n", command, payload)
	}
	return err
}

// StartHeartbeat 启动心跳
func (c *Client) StartHeartbeat() (err error) {
	// 启动心跳协程
	go func() {
		ticker := time.NewTicker(time.Second * global.HeartbeatInterval)
		// 创建心跳定时器
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// 检查连接状态
				if c.Status != enums.ClientStatusConnected {
					err = fmt.Errorf("客户端未连接，停止心跳发送")
				}
				// 发送心跳包
				err = c.Handler.HeartbeatReq()

				if err != nil {
					err = fmt.Errorf("发送心跳包失败，Error: %v\n", err)
				}
			}
		}
	}()
	return err
}
