package core

import (
	"context"
	"time"

	config "yunyez/internal/common/config"
	logger "yunyez/internal/pkg/logger"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var (
	mqtt_address = config.GetString("mqtt.address") // MQTT 代理地址
	username = config.GetString("mqtt.client.username") // MQTT 客户端用户名
	password = config.GetString("mqtt.client.password") // MQTT 客户端密码

	MQTT_CLIENT paho.Client
)


var (
	// OnConnectHandler 连接成功处理函数
	// 当客户端成功连接到 MQTT 代理时调用
	OnConnectHandler paho.OnConnectHandler = func(client paho.Client) {
		logger.Info(context.TODO(), "mqtt.connect", map[string]interface{}{
			"address": mqtt_address,
			
		})
	}

	// ConnectionLostHandler 连接丢失处理函数
	// 当客户端与 MQTT 代理断开连接时调用
	ConnectionLostHandler paho.ConnectionLostHandler = func(client paho.Client, err error) {
		logger.Error(context.TODO(), "mqtt.connection_lost", map[string]interface{}{
			"error": err,
		})
	}
)

func InitMQTTClient() error {
	opts := paho.NewClientOptions().AddBroker(mqtt_address)
	// 其他配置
	opts.SetClientID(uuid.New().String())
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetKeepAlive(60 * time.Second) // 60秒发送一次心跳包
	opts.SetPingTimeout(5 * time.Second) // 5秒未收到心跳包回复，认为连接已断开
	opts.SetProtocolVersion(4) // MQTT 3.1.1 协议版本
	opts.SetAutoReconnect(true) // 自动重连
	opts.SetOnConnectHandler(OnConnectHandler)
	opts.SetConnectionLostHandler(ConnectionLostHandler)

	MQTT_CLIENT = paho.NewClient(opts)
	// 连接MQTT集群
	if err := ConnectToBroker(MQTT_CLIENT); err != nil {
		return err
	}

	// 健康检测
	// 每30秒检测一次连接状态，若断开则尝试重连
	go func() {
		for {
			if !MQTT_CLIENT.IsConnected() {
				if err := ConnectToBroker(MQTT_CLIENT); err != nil {
					logger.Error(context.TODO(), "mqtt.connect error", map[string]interface{}{
						"error": err,
					})
				}
			}
			time.Sleep(30 * time.Second)
		}
	}()

	logger.Info(context.TODO(), "mqtt.client init success", map[string]interface{}{
		"address": mqtt_address,
		"client_id": opts.ClientID,
	})
	return nil
}

// ConnectToBroker 连接到 MQTT 代理
// 参数：
//   - client: MQTT 客户端实例
// 返回值:
//   - error: 连接错误信息，若连接成功则为 nil
func ConnectToBroker(client paho.Client) error {
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}
