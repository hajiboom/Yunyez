package handler


// 转发mqtt消息到其他协议接口
// 使用自己封装的mqtt-bridge 来实现消息转发：https://github.com/hajiboom/mqtt-bridge
// 包括但不限于：
// - http接口
// - 消息队列
// - 其他协议转换

import (

	bridge "github.com/hajiboom/mqtt-bridge"
)
 

type Forward struct {
	Bridge *bridge.Bridge
}

