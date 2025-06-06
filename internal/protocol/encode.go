package protocol

import (
	"bytes"
	"encoding/binary"
)

// Encode 自定义协议消息编码
func Encode(data []byte) ([]byte, error) {
	// 1. 消息头：消息长度（4字节）
	length := int32(len(data))
	// 向系统为具有读写方法的字节大小可变的缓冲区申请内存
	pkg := new(bytes.Buffer)

	// 2. 写入消息头
	err := binary.Write(pkg, binary.LittleEndian, length)

	if err != nil {
		return nil, err
	}

	// 3. 写入消息体
	err = binary.Write(pkg, binary.LittleEndian, data)
	if err != nil {
		return nil, err
	}
	// 4.返回封包完毕的缓冲区中数据
	return pkg.Bytes(), nil
}
