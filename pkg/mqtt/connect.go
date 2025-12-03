package mqtt

// ======================== MQTT连接 ======================================
// mqtt生命周期管理
// 业务处理回调



// StartConnect 启动MQTT连接
func StartConnect() error {
	// 初始化mqtt中间件
	// 创建监听mqtt topic的客户端
	// 健康检查 自动重连
	// 保持主程序运行
	select {}
}






// ========================= MQTT桥接转发 ======================================
// 桥接器转发topic消息到指定接口
// 该中间件还在开发中，目前只支持转发到HTTP接口，后续会支持其他协议类型，欢迎参与开发
// link：https://github.com/hajiboom/mqtt-bridge



// StartBridge 启动MQTT桥接转发
func StartBridge() error {
	// 初始化mqtt桥接器
	// 添加中间件链处理 【TODO】
	// 订阅需要桥接的topic 会根据配置自动转发消息
	// 保持主程序运行
	select {}
}


