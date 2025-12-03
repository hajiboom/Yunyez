// topic.go 定制化 MQTT 主题
package core

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"yunyez/internal/common/constant"
	"yunyez/pkg/logger"

	_ "github.com/eclipse/paho.mqtt.golang"
)

const(
	MQTT_TOPIC_LEVEL = 5 // mqtt topic 层级
	DEVICE_SN_REGEX = "^[a-zA-Z0-9_]+$" // 设备序列号格式

)

var (
	regDeviceSN = regexp.MustCompile(DEVICE_SN_REGEX)
)

// Topic 用于表示 MQTT 主题
// 基本结构：
// <厂商名称>/<设备类型>/<设备序列号>/<命令类型>/<标识>
type Topic struct {
	Vendor      string // 厂商名称
	DeviceType  string // 设备类型
	DeviceSN    string // 设备序列号
	CommandType string // 命令类型
	Flag        string `validate:"oneof=server client"` // 标识（server / client）用于区分上传下发
}

// String 实现 fmt.Stringer 接口，用于将 Topic 转换为字符串
func (t *Topic) String() string {
	if t == nil {
		return ""
	}
	return strings.Join([]string{
		t.Vendor,
		t.DeviceType,
		t.DeviceSN,
		t.CommandType,
		t.Flag,
	}, "/")
}

// 校验 Topic 是否合法
func (t *Topic) Validate() error {
	// 检查vendor是否是合法厂商名称
	if !validateVendor(t.Vendor) {
		return fmt.Errorf("vendor %s is invalid", t.Vendor)
	}
	// 检查设备类型，设备序列号是否是合法格式
	if !validateDeviceSN(t.DeviceSN) {
		return fmt.Errorf("deviceSN %s is invalid", t.DeviceSN)
	}
	// 检查命令类型是否是合法格式
	if t.CommandType == "" {
		return fmt.Errorf("commandType is empty")
	}
	// 检查标识是否是 server 或 client
	if t.Flag == "" {
		return fmt.Errorf("flag is empty")
	}
	return nil
}

// TopicParse 解析Topic
func TopicParse(topic string) (*Topic, error) {
	if topic == "" {
		logger.Error(context.Background(), "topic is empty", nil)
		return nil, fmt.Errorf("topic is empty")
	}
	parts := strings.Split(topic, "/")
	if len(parts) < MQTT_TOPIC_LEVEL {
		logger.Error(context.Background(), "topic %s is invalid", map[string]any{
			"topic": topic,
			"length": len(parts),
		})
		return nil, fmt.Errorf("topic %s is invalid", topic)
	}
	obj := &Topic{
		Vendor:      parts[0],
		DeviceType:  parts[1],
		DeviceSN:    parts[2],
		CommandType: parts[3],
		Flag:        parts[4],
	}
	if err := obj.Validate(); err != nil {
		return nil, err
	}
	return obj, nil
}


// ValidateVendor 校验厂商名称是否是合法厂商名称
func validateVendor(vendor string) bool {
	if vendor == "" {
		return false
	}
	if _, ok := constant.GetVendor()[vendor]; !ok {
		return false
	}
	return true
}


// ValidateDeviceSN 校验设备序列号是否是合法
// 设备序列号格式：
func validateDeviceSN(deviceSN string) bool {
	if deviceSN == "" {
		return false
	}
	if !regDeviceSN.MatchString(deviceSN) {
		return false
	}
	return true
}
