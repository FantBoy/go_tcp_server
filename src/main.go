package main

import (
	"fmt"
	. "github.com/FantBoy/go_tcp_server/src/libs"
	"github.com/pkg/errors"
)

func server() error {
	endpoint := NewTcpServer()

	endpoint.AddHandleFunc("string", HandleStrings)
	endpoint.AddHandleFunc("gob", HandleGob)

	// 开始监听
	return endpoint.Listen()
}

func main() {
	err := server()
	if err != nil {
		fmt.Println("Error:", errors.WithStack(err))
	}
}
