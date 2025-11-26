package main

import (
	"log"
	"yunyez/internal/common/config"
)

// 设备服务
// 设备通信支持：mqtt，websocket

func main() {

	name := config.GetString("name")
	log.Printf("name: %s", name)
}