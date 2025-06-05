package serializer

import (
	"bufio"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"tcpsocketv2/global"
	"tcpsocketv2/intranal/protocol"
	message "tcpsocketv2/pb"
	"tcpsocketv2/pkg/utils"
)

// SerializeMessage 序列化消息
func SerializeMessage(command message.CommandType, payload proto.Message) ([]byte, error) {
	// 包装 payload
	payloadAny, err := anypb.New(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	timestamp := utils.GetCurrentTimestamp()
	// 构建 消息体
	msgBody := &message.MSG_BODY{
		Command:   command.Enum(),
		Payload:   payloadAny,
		Timestamp: &timestamp,
	}
	// 序列化 消息体
	msgBodyBytes, marshalErr := proto.Marshal(msgBody)
	if marshalErr != nil {
		return nil, fmt.Errorf("failed to marshal message body: %v", marshalErr)
	}

	// 对消息体进行编码
	pkg, _ := protocol.Encode(msgBodyBytes)
	return pkg, nil
}

// DeserializeMessage 反序列化消息
func DeserializeMessage(reader *bufio.Reader) (message.CommandType, proto.Message, error) {
	// 先进行协议解码
	decodedData, err := protocol.Decode(reader)
	if err == io.EOF {
		fmt.Printf("Receive Over!\n")
		return 0, nil, err
	}
	if err != nil {
		return 0, nil, fmt.Errorf("failed to decode data: %v", err)
	}

	// 反序列化消息体
	msgBody := &message.MSG_BODY{}
	if err := proto.Unmarshal(decodedData, msgBody); err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal message body: %v", err)
	}
	timestamp := msgBody.GetTimestamp()
	expireTime := utils.GetCurrentTimestamp() - global.MsgExpireTime
	// 判断消息是否过期
	if timestamp < expireTime {
		return 0, nil, fmt.Errorf("消息非法，超过过期时间，当前时间%v, 消息时间: %v\n", utils.GetCurrentTimestamp(), timestamp)
	}
	command := msgBody.GetCommand()
	payload := msgBody.GetPayload()
	if payload == nil {
		return command, nil, fmt.Errorf("payload is nil")
	}

	// 根据不同的消息类型创建对应的 payload 结构
	var payloadMsg proto.Message
	fmt.Printf("收到指令类型为： %v\n", command)
	switch command {
	case message.CommandType_CommandType_HandShakeReq:
		fmt.Println("收到指令：握手请求消息")
		payloadMsg = &message.MSG_HANDSHAKE_REQ{}
	case message.CommandType_CommandType_HandShakeResp:
		fmt.Println("收到指令：握手响应消息")
		payloadMsg = &message.MSG_HANDSHAKE_RESP{}
	case message.CommandType_CommandType_Heartbeat:
		fmt.Println("收到指令：心跳消息")
		payloadMsg = &message.MSG_HEARTBEAT{}
	case message.CommandType_CommandType_Unknow:
		fmt.Println("收到指令：未知消息")
	// 添加更多 case 处理其他命令类型
	default:
		return command, nil, fmt.Errorf("unsupported handler type: %v", command)
	}
	if err := payload.UnmarshalTo(payloadMsg); err != nil {
		return command, nil, fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	// 返回解析后的命令类型、对应的消息体以及错误为nil
	return command, payloadMsg, nil
}
