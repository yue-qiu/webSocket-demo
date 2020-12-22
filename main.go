package main

import (
	"fmt"
	"github.com/yue-qiu/go-WebSocket/src"
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Panic(err.Error())
	}

	fmt.Printf("Running")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept Error: %v\n", err.Error())
			continue
		}

		go src.HandleWSocketConn(conn)

	}
}
