package main

import (
	_ "yunyez/internal/common/config"
	device "yunyez/internal/app/device"
)



// 设备服务
// 设备通信支持：mqtt，websocket

func main() {
	// 启动设备后端服务
	device.Start()
}