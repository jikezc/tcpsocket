package socket

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"tcpsocketv2/global"
	"tcpsocketv2/intranal/serializer"
	message "tcpsocketv2/pb"
	"tcpsocketv2/pkg/utils"
	"time"
)

// Spec 客户端硬件信息
type Spec struct {
	Os  string
	Cpu float64
	Mem float64
}

// Session 会话信息
type Session struct {
	DiverId       string // 设备ID
	LastAliveTime int64  //  最后活跃时间
	ClientSpec    Spec   // 客户端硬件信息
}

// Handler 接口：处理消息
type Handler interface {
	HandlerHandshake(conn net.Conn, payload *message.MSG_HANDSHAKE_REQ) error
	HandlerHeartbeat(conn net.Conn, payload *message.MSG_HEARTBEAT) error
}

// Server TCP 服务器
type Server struct {
	Address    string               // 监听地址
	Port       int                  // 监听端口
	SessionMap map[net.Conn]Session // 会话连接池
	Handler    Handler              // 消息处理器
}

// NewServer 创建并返回一个Server实例，并初始化SessionMap
func NewServer(address string, port int) Server {
	return Server{
		Address:    address,
		Port:       port,
		SessionMap: make(map[net.Conn]Session), // 使用make函数初始化map
		Handler:    nil,
	}
}

// RegisterHandler 注册消息处理器
func (s *Server) RegisterHandler(handler Handler) {
	s.Handler = handler
}

// GetSession 获取指定连接的Session信息
func (s *Server) GetSession(conn net.Conn) (Session, bool) {
	_session, ok := s.SessionMap[conn]
	return _session, ok
}

// UpdateSession 更新Session信息
func (s *Server) UpdateSession(conn net.Conn, _session Session) {
	s.SessionMap[conn] = _session
}

// ListenAndServe 启动TCP服务器并开始监听
func (s *Server) ListenAndServe() error {
	server := fmt.Sprintf("%s:%d", s.Address, s.Port)
	listener, err := net.Listen("tcp", server)
	if err != nil {
		return fmt.Errorf("Start TCP Server on %v Failed\nerr: %v", server, err)
	}
	fmt.Printf("Server Listening: %v\n", server)
	defer func() {
		err := listener.Close()
		if err != nil {
			fmt.Printf("Listen Close Error: %v\n", err)
		}
	}()

	for {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			fmt.Printf("Accept Error: %v\n", acceptErr)
			continue
		}
		fmt.Printf("Accept Client Conn: %v\n", conn.RemoteAddr())
		// 启动一个goroutine处理连接
		go s.handleConnection(conn)
	}
}

// handleConnection 处理客户端连接
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Conn Close Error: %v\n", err)
		}
	}()
	reader := bufio.NewReader(conn)

	for {
		clientIP := conn.RemoteAddr()
		// 反序列化消息
		command, payload, err := serializer.DeserializeMessage(reader)
		// 消息结束符则不再继续
		if err == io.EOF {
			break
		}
		// 反序列化失败
		if err != nil {
			fmt.Printf("Client: %s, DeserializeMessage Error: %v\n", clientIP, err)
			break
		}
		fmt.Printf("Client: %s, Receive Message: %v\n", clientIP, payload)

		// 处理消息
		handlerErr := s.handleMessage(conn, command, payload)
		if handlerErr != nil {
			fmt.Printf("Client: %s, HandleMessage Error: %v\n", clientIP, err)
			break
		}
	}
}

// handleMessage 处理接收到的消息，根据命令类型调用对应的处理器函数
func (s *Server) handleMessage(conn net.Conn, command message.CommandType, payload interface{}) (err error) {
	switch command {
	case message.CommandType_CommandType_HandShakeReq:
		err = s.Handler.HandlerHandshake(conn, payload.(*message.MSG_HANDSHAKE_REQ))
	case message.CommandType_CommandType_Heartbeat:
		err = s.Handler.HandlerHeartbeat(conn, payload.(*message.MSG_HEARTBEAT))
	default:
		fmt.Printf("收到未知指令: %v\n", command)
	}

	if err != nil {
		fmt.Printf("Server, 处理消息异常: %v\n", err)
	}
	return err
}

// StartHeartbeatChecker 启动心跳检查协程
func (s *Server) StartHeartbeatChecker(conn net.Conn) {
	// 获取停止通道用于管理心跳协程生命周期
	checkHeartbeat(s, conn)
	// 可以在这里记录日志或添加清理逻辑
	fmt.Printf("Heartbeat checker started for client: %v\n", conn.RemoteAddr())
	//defer func() {
	//	// 停止心跳检查
	//	stopC <- true
	//	fmt.Printf("Heartbeat checker stopped for client: %v\n", conn.RemoteAddr())
	//}()
}

// checkHeartbeat 检测心跳
func checkHeartbeat(s *Server, conn net.Conn) (stopC chan bool) {
	ticker := time.NewTicker(global.HeartbeatCheckTime * time.Second)
	stopC = make(chan bool)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_session, ok := s.SessionMap[conn]
				if !ok {
					fmt.Println("客户端会话不存在，停止心跳检测")
					return
				}
				// 检查心跳超时
				current := utils.GetCurrentTimestamp()
				interval := current - _session.LastAliveTime
				if interval > global.HeartbeatTimeout {
					fmt.Printf("客户端: %v, 心跳超时，关闭连接\n", conn.RemoteAddr())
					delete(s.SessionMap, conn)
					err := conn.Close()
					if err != nil {
						fmt.Printf("Server, 关闭连接异常: %v\n", err)
					}
					return
				} else {
					fmt.Printf("客户端: %v, 心跳正常\n", conn.RemoteAddr())
				}
			case <-stopC:
				fmt.Println("停止心跳检查")
				return
			}
		}
	}()
	return stopC
}
