package src

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strings"
)

const salt = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func HandleWSocketConn(conn net.Conn) {
	raw := make([]byte, 1024)
	_, err := conn.Read(raw)
	log.Println(string(raw))
	if err != nil {
		log.Println(err.Error())
		return
	}

	// 握手默认使用 GET 方法
	if string(raw[: 3]) != "GET" {
		log.Printf("HandleWsocketConn Error: Must be GET method")
		return
	}

	header := parseHTTPHandShakeHeader(raw)

	// 握手协议头字段错误
	if header["Upgrade"] != "websocket" || header["Connection"] != "Upgrade" || header["Sec-WebSocket-Key"] == "" {
		log.Println("Header Format Error")
		return
	}

	// 得到响应 Sec-WebSocket-Accept 值
	sha := sha1.New()
	sha.Write([]byte(header["Sec-WebSocket-Key"]+salt))
	accept := make([]byte, 28)
	base64.StdEncoding.Encode(accept, sha.Sum(nil))

	log.Printf("Key: %v, Accept: %v\n", header["Sec-WebSocket-Key"], string(accept))

	rsp := GetHTTPHandShakeRsp(accept)

	// 发送响应报文，握手成功
	if _, err = conn.Write([]byte(rsp)); err != nil {
		log.Println(err.Error())
		return
	}

	wSocket := NewWSocket(conn)

	for {
		data, err := wSocket.Read()
		if err != nil {
			log.Printf("wSocket Read Error: %v\n", err.Error())
			wSocket.Conn.Close()
			return
		}

		log.Printf("Read data from wSocket: %v\n", string(data))

		wSocket.Write([]byte(fmt.Sprintf("%500v\n", "hello")))
		log.Println("send data")
	}
}

// 解析基于 HTTP 的握手请求报文，提取头部
func parseHTTPHandShakeHeader(raw []byte) map[string]string {
	data := string(raw)
	lines := strings.Split(data, "\r\n")
	header := make(map[string]string)

	for _, line := range lines {
		words := strings.Split(line, ":")
		if len(words) == 2 {
			header[strings.Trim(words[0], " ")] = strings.Trim(words[1], " ")
		}
	}

	return header
}

// 构造 HTTP 握手响应报文
func GetHTTPHandShakeRsp(accept []byte) string {
	rsp := "HTTP/1.1 101 Switching Protocols\r\n"
	rsp += "Upgrade: websocket\r\n"
	rsp += "Connection: Upgrade\r\n"
	rsp += "Sec-WebSocket-Accept: " + string(accept) + "\r\n"

	rsp += "\r\n"

	return rsp
}