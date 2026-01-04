package core

import (
	"context"
	"time"

	config "yunyez/internal/common/config"
	logger "yunyez/internal/pkg/logger"
	"yunyez/internal/pkg/mqtt/handler"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var (
	mqttAddress = config.GetString("mqtt.address")         // MQTT 代理地址
	username    = config.GetString("mqtt.client.username") // MQTT 客户端用户名
	password    = config.GetString("mqtt.client.password") // MQTT 客户端密码
	Model       = config.GetString("rule.model")           // 消息转发模型

	MqttClient paho.Client // 原生MQTT客户端实例
)

var (
	// OnConnectHandler 连接成功处理函数
	// 当客户端成功连接到 MQTT 代理时调用
	// 1.订阅主题
	OnConnectHandler paho.OnConnectHandler = func(client paho.Client) {
		topics := config.GetList("topics")
		if len(topics) == 0 {
			logger.Error(context.TODO(), "mqtt.core.OnConnectHandler error", map[string]interface{}{
				"error": "rule.topics is empty",
			})
			return
		}
		// 订阅主题
		err := SubscribeTopic(client, topics)
		if err != nil {
			logger.Error(context.TODO(), "mqtt.core.OnConnectHandler error", map[string]interface{}{
				"error": err,
			})
			return
		}
	}

	// ConnectionLostHandler 连接丢失处理函数
	// 当客户端与 MQTT 代理断开连接时调用
	ConnectionLostHandler paho.ConnectionLostHandler = func(client paho.Client, err error) {
		logger.Error(context.TODO(), "mqtt.connection_lost", map[string]interface{}{
			"error": err,
		})
	}

	// MessageHandler 默认消息处理函数
	// 当客户端订阅的主题有新消息到达时调用
	// 1. 消息预处理 @deprecated 暂时不处理
	// 2. 消息路由
	MessageHandler paho.MessageHandler = func(client paho.Client, msg paho.Message) {
		go func() { // 异步处理消息防止阻塞
			ctx := context.Background()
			// 转发消息
			topic := msg.Topic()
			topicObj, err := TopicParse(topic)
			if err != nil {
				logger.Error(context.TODO(), "mqtt.core.TopicParse error", map[string]interface{}{
					"topic": topic,
					"error": err,
				})
				return
			}

			message := &handler.Message{
				Topic:       topic,
				CommandType: topicObj.CommandType,
				ClientID:    topicObj.DeviceSN,
				Content:     msg.Payload(),
				StartTime:   time.Now().Format("2006-01-02 15:04:05"),
			}

			// 1.路由策略
			strategy := &handler.SendHandler{}
			strategy.Set(Model, &MqttClient)
			strategy.Send(ctx, message)

			logger.Info(ctx, "mqtt.core.MessageHandler success", map[string]interface{}{
				"topic": topic,
			})
		}()
	}
)

func InitMQTTClient() error {
	opts := paho.NewClientOptions().AddBroker(mqttAddress)
	// 其他配置
	opts.SetClientID(uuid.New().String())
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetKeepAlive(60 * time.Second)                  // 60秒发送一次心跳包
	opts.SetPingTimeout(5 * time.Second)                 // 5秒未收到心跳包回复，认为连接已断开
	opts.SetProtocolVersion(4)                           // MQTT 3.1.1 协议版本
	opts.SetAutoReconnect(true)                          // 自动重连
	opts.SetOnConnectHandler(OnConnectHandler)           // 连接成功处理函数
	opts.SetConnectionLostHandler(ConnectionLostHandler) // 连接丢失处理函数
	opts.SetDefaultPublishHandler(MessageHandler)        // 默认消息处理函数

	MqttClient = paho.NewClient(opts)
	// 连接MQTT集群
	if err := ConnectToBroker(MqttClient); err != nil {
		return err
	}

	// 健康检测
	// 每30秒检测一次连接状态，若断开则尝试重连
	go func() {
		for {
			if !MqttClient.IsConnected() {
				if err := ConnectToBroker(MqttClient); err != nil {
					logger.Error(context.TODO(), "mqtt.connect error", map[string]interface{}{
						"error": err,
					})
				}
			}
			time.Sleep(30 * time.Second)
		}
	}()

	logger.Info(context.TODO(), "mqtt.client init success", map[string]interface{}{
		"address":   mqttAddress,
		"client_id": opts.ClientID,
	})
	return nil
}

// ConnectToBroker 连接到 MQTT 代理
// 参数：
//   - client: MQTT 客户端实例
//
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
