package src

import (
	"log"
	"net"
)

type WSocket struct {
	Conn net.Conn
	Mask []byte
}

func NewWSocket(conn net.Conn) *WSocket {
	return &WSocket{
		Conn: conn,
	}
}

func (w *WSocket) Read() (data []byte, err error) {
	flagByte := make([]byte, 1)
	_, err = w.Conn.Read(flagByte)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	FIN := (flagByte[0] >> 7) & 1
	RSV1 := (flagByte[0] >> 6) & 1
	RSV2 := (flagByte[0] >> 5) & 1
	RSV3 := (flagByte[0] >> 4) & 1
	opcode := flagByte[0] & 15
	log.Printf("FIN: %v, RSV1: %v, RSV2: %v, RSV3: %v, opcode: %x\n", FIN, RSV1, RSV2, RSV3, opcode)

	payloadLenByte := make([]byte, 1)
	_, err = w.Conn.Read(payloadLenByte)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	MASK := (payloadLenByte[0] >> 7) & 1
	payloadLen := int(payloadLenByte[0] & 0x7f)
	if payloadLen == 126 {
		extended2Byte := make([]byte, 2)
		w.Conn.Read(extended2Byte)
	} else if payloadLen == 127 {
		extended8Byte := make([]byte, 8)
		w.Conn.Read(extended8Byte)
	}

	log.Printf("MASK: %v, payloadLen: %v\n", MASK, payloadLen)

	// 读掩码
	if MASK == 1 {
		maskByte := make([]byte, 4)
		w.Conn.Read(maskByte)
		w.Mask = maskByte
	}

	payloadDataByte := make([]byte, payloadLen)
	_, err = w.Conn.Read(payloadDataByte)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	data = make([]byte, payloadLen)

	// 用掩码解析数据
	for i := 0; i < payloadLen; i++ {
		if MASK == 1 {
			data[i] = payloadDataByte[i] ^ w.Mask[i % 4]
		} else {
			data[i] = payloadDataByte[i]
		}
	}

	// 一个消息最后的片段，不必再等待拼接了
	if FIN == 1 {
		return
	}

	// 等待更多片段过来组成完整数据
	nextData, err := w.Read()
	if err != nil {
		return
	}

	data = append(data, nextData...)

	return data, nil
}

func (w *WSocket) Write(raw []byte) {
	if len(raw) < 126 {
		w.sendAllData(raw)
	} else {
		w.sendFragData(raw)
	}
}

func (w *WSocket) sendAllData(data []byte)  {
	w.Conn.Write([]byte{0x81})

	payLoadLenByte := byte(0x00) | byte(len(data))
	w.Conn.Write([]byte{payLoadLenByte})

	w.Conn.Write(data)
}

func (w *WSocket) sendFragData(raw []byte) {
	// 发送分段帧首帧
	w.Conn.Write([]byte{0x01})

	w.Conn.Write([]byte{0x7d})

	w.Conn.Write(raw[: 125])

	// 发送分段帧中间帧
	for {
		raw = raw[125: ]
		if len(raw) <= 125 {
			break
		}

		w.Conn.Write([]byte{0x00})

		w.Conn.Write([]byte{0x7d})

		w.Conn.Write(raw[: 125])
	}

	// 发送分段帧尾帧
	w.Conn.Write([]byte{0x80})

	w.Conn.Write([]byte{byte(len(raw))})

	w.Conn.Write(raw)
}
