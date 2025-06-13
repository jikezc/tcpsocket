package socket

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"tcpsocketv2/common/logger"
	"tcpsocketv2/config"
	"tcpsocketv2/internal/serializer"
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
	DiverId       string          // 设备ID
	LastAliveTime int64           //  最后活跃时间
	ClientSpec    Spec            // 客户端硬件信息
	Ctx           context.Context // 会话上下文
}

// ServMsgHandlerInterface 接口：处理消息
type ServMsgHandlerInterface interface {
	HandleHandshakeReq(conn net.Conn, payload *message.MSG_HANDSHAKE_REQ, ctx context.Context) error
	HandleHeartbeatReq(conn net.Conn, payload *message.MSG_HEARTBEAT) error
}

// Server TCP 服务器
type Server struct {
	Address    string                  // 监听地址
	Port       int                     // 监听端口
	SessionMap map[net.Conn]Session    // 会话连接池
	Handler    ServMsgHandlerInterface // 消息处理器
}

// NewServer 创建并返回一个Server实例，并初始化SessionMap
func NewServer(address string, port int) *Server {
	return &Server{
		Address:    address,
		Port:       port,
		SessionMap: make(map[net.Conn]Session), // 使用make函数初始化map
		Handler:    nil,
	}
}

// RegisterHandler 注册消息处理器
func (s *Server) RegisterHandler(handler ServMsgHandlerInterface) {
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
	l := logger.Get()
	server := fmt.Sprintf("%s:%d", s.Address, s.Port)
	listener, err := net.Listen("tcp", server)
	if err != nil {
		return fmt.Errorf("Start TCP Server on %v Failed\nerr: %v", server, err)
	}
	l.Info(fmt.Sprintf("Server Listening: %s ", server))
	defer func() {
		err := listener.Close()
		if err != nil {
			l.Error(fmt.Sprintf("Listen Close Error: %v\n", err))
		}
	}()

	for {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			l.Error(fmt.Sprintf("Accept Error: %v", acceptErr))
			continue
		}
		l.Info(fmt.Sprintf("Accept Client Conn: %v", conn.RemoteAddr()))

		// 启动一个goroutine处理连接
		go s.handleConnection(conn)

	}
}

// handleConnection 处理客户端连接
func (s *Server) handleConnection(conn net.Conn) {
	// 初始化上下文，用于传递必要参数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 获取日志实例, 并将上下文传递给日志实例
	l := logger.Get()
	ctx = logger.WithCtx(ctx, l)

	defer func() {
		err := conn.Close()
		if err != nil {
			l.Error(fmt.Sprintf("Close Client Conn Error: %v\n", err))
		}
	}()
	reader := bufio.NewReader(conn)
	clientIp := conn.RemoteAddr().String()
	for {
		// 反序列化消息
		command, payload, err := serializer.DeserializeMessage(reader, ctx)
		// 消息结束符则不再继续
		if err == io.EOF {
			break
		}
		// 反序列化失败
		if err != nil {
			l.Error(fmt.Sprintf("Server DeserializeMessage Error: %v, Client: %s", err, clientIp))
			break
		}
		l.Info(fmt.Sprintf("Server Receive Message: %v, Client: %s", payload, clientIp))
		// 处理消息
		handlerErr := s.handleMessage(command, payload, conn, ctx)
		if handlerErr != nil {
			l.Error(fmt.Sprintf("Server HandleMessage Error: %v, Client: %s", err, clientIp))
			break
		}
	}
}

// handleMessage 处理接收到的消息，根据命令类型调用对应的处理器函数
func (s *Server) handleMessage(command message.CommandType, payload interface{}, conn net.Conn, ctx context.Context) (err error) {
	l := logger.FromCtx(ctx)
	switch command {
	case message.CommandType_CommandType_HandShakeReq:
		err = s.Handler.HandleHandshakeReq(conn, payload.(*message.MSG_HANDSHAKE_REQ), ctx)
	case message.CommandType_CommandType_Heartbeat:
		err = s.Handler.HandleHeartbeatReq(conn, payload.(*message.MSG_HEARTBEAT))
	default:
		l.Warn(fmt.Sprintf("收到未知指令: %v\n", command))
	}

	if err != nil {
		l.Error(fmt.Sprintf("Server, 处理消息异常: %v\n", err))
	}
	return err
}

// StartHeartbeatChecker 启动心跳检查协程
func (s *Server) StartHeartbeatChecker(conn net.Conn) {
	ctx := s.SessionMap[conn].Ctx
	l := logger.FromCtx(ctx)
	// 获取停止通道用于管理心跳协程生命周期
	checkHeartbeat(s, conn)
	// 可以在这里记录日志或添加清理逻辑
	l.Debug(fmt.Sprintf("Heartbeat checker started for client: %v", conn.RemoteAddr()))
}

// checkHeartbeat 检测心跳
func checkHeartbeat(s *Server, conn net.Conn) (stopC chan bool) {
	cfg := config.Get()
	ctx := s.SessionMap[conn].Ctx
	l := logger.FromCtx(ctx)
	ticker := time.NewTicker(cfg.Msg.HeartbeatCheckTime * time.Second)
	stopC = make(chan bool)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_session, ok := s.SessionMap[conn]
				if !ok {
					l.Error("客户端会话不存在，停止心跳检测")
					return
				}
				// 检查心跳超时
				current := utils.GetCurrentTimestamp()
				interval := current - _session.LastAliveTime
				if interval > cfg.Msg.HeartbeatTimeout {
					l.Error(fmt.Sprintf("客户端: %v, 心跳超时，关闭连接", conn.RemoteAddr()))
					delete(s.SessionMap, conn)
					err := conn.Close()
					if err != nil {
						l.Error(fmt.Sprintf("Server, 关闭连接异常: %v\n", err))
					}
					return
				} else {
					l.Debug(fmt.Sprintf("客户端: %v, 心跳正常", conn.RemoteAddr()))
				}
			case <-stopC:
				l.Warn("停止心跳检查")
				return
			}
		}
	}()
	return stopC
}
