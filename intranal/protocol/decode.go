package protocol

import (
	"bufio"
	"bytes"
	"encoding/binary"
)

// Decode 自定义协议消息解码
func Decode(reader *bufio.Reader) ([]byte, error) {
	// 1. 读取头部信息，获取数据长度
	lengthByte, err := reader.Peek(4)
	if err != nil {
		return nil, err
	}
	lengthBuffer := bytes.NewBuffer(lengthByte)
	var length int32
	err = binary.Read(lengthBuffer, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}
	// 2. Buffered 返回缓冲区中现有的可读字节数，如果获取的字节数小于消息的长度则说明数据包有误
	if int32(reader.Buffered()) < length+4 {
		return nil, err
	}

	// 3. 读取消息内容
	pack := make([]byte, int(4+length))
	_, err = reader.Read(pack)
	if err != nil {
		return nil, err
	}
	return pack[4:], nil
}
