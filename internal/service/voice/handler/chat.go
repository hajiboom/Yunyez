// Package handler provides voice processing services pipeline.
// including:
// - voice: the voice processing service
// - asr: the voice recognition service
// - nlu: the natural language understanding service
// - chat: the natural language conversation service
// - tts: the text-to-speech service
package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
	config "yunyez/internal/common/config"
	constant "yunyez/internal/common/constant"
	tools "yunyez/internal/common/tools"
	asr "yunyez/internal/pkg/agent/asr"
	llm "yunyez/internal/pkg/agent/llm"
	metering "yunyez/internal/pkg/agent/metering"
	nlu "yunyez/internal/pkg/agent/nlu"
	tts "yunyez/internal/pkg/agent/tts"
	logger "yunyez/internal/pkg/logger"
	mqttCommon "yunyez/internal/pkg/mqtt/common"
	mqttCore "yunyez/internal/pkg/mqtt/core"
	voice "yunyez/internal/pkg/mqtt/protocol/voice"
	buffer "yunyez/internal/service/voice/buffer"
)

var (
	asrModel  = config.GetString("asr.model")   // default asr model
	nluModel  = config.GetString("nlu.model")   // default nlu model
	chatModel = config.GetString("agent.model") // default chat model： qwen
	ttsModel  = config.GetString("tts.model")   // default tts model

	asrEndpoint = config.GetString("asr.endpoint") // default asr endpoint
	nluEndpoint = config.GetString("nlu.endpoint") // default nlu endpoint

	agentStrategy *llm.Strategy // llm model strategy

	asrClient asr.Service
	nluClient nlu.Client
	ttsClient tts.Service

	publishCounter uint64 // TODO: 测试暂存序列号 记得移除
)

// 第三方api成本计算
var (
	model = config.GetString(fmt.Sprintf("%s.model", chatModel)) // qwen-flash
	rules = map[string]metering.PricingRule{
		model: {
			InputPrice:  config.GetFloat64WithDefault("pricing.models.qwen-flash.input_price", 0.001),
			OutputPrice: config.GetFloat64WithDefault("pricing.models.qwen-flash.output_price", 0.002),
			Currency:    config.GetStringWithDefault("pricing.models.qwen-flash.currency", "CNY"),
		},
	}
	meteringService *metering.MeteringService // metering service
)

func init() {
	// 初始化智能体模型
	asrClient = asr.NewASRClient(asrModel, asrEndpoint) // init asr client
	nluClient = nlu.NewClient(nluEndpoint)              // init nlu client

	agentStrategy = &llm.Strategy{}
	agentStrategy.SetAgent(chatModel)

	ttsClient = tts.NewTTSClient() // init tts client
	// init metering service
	meteringService = metering.Initialize(rules)
}

// ChatPipeline response the natural language conversation response
// step:
// 1. asr: recognize the voice message to text
// 2. nlu: understand the text message to intent and entities
// 3. chat: call the llm model to chat
// 4. tts: generate the audio data from the chat response by tts
// 5. record: record the cost into metering repository if exists
// Parameters:
//   - ctx: the context.Context object
//   - clientID: the device sequence number to generate the MQTT topic
//   - message: the audio message from mqtt
//
// Returns:
//   - error: the error object if the chat failed
func ChatPipeline(ctx context.Context, clientID string, message []byte) error {
	if message == nil {
		return fmt.Errorf("message is nil")
	}

	// asr
	text, err := asrClient.Transfer(ctx, message)
	if err != nil {
		logger.Error(ctx, "asr service failed", map[string]any{
			"error":     err.Error(),
			"clientID":  clientID,
			"audio_len": len(message),
		})
		return err
	}
	// nlu
	intent, err := nluClient.Predict(&nlu.Input{
		Text: text,
	})
	if err != nil {
		logger.Error(ctx, "nlu service failed", map[string]any{
			"error":    err.Error(),
			"clientID": clientID,
			"text":     text,
		})
		return err
	}

	// judge intent from input text
	if intent.Intent != constant.IntentChitChat {
		// special intent command
		err := SpecialAction(ctx, clientID, intent)
		if err != nil {
			logger.Error(ctx, "special action failed", map[string]any{
				"error":             err.Error(),
				"clientID":          clientID,
				"text":              text,
				"intent":            intent.Intent,
				"intent_confidence": intent.Confidence,
				"intent_is_command": intent.IsCommand,
			})
			return err
		}
		return nil
	}

	// chat -- call the llm model to response in streaming
	replyChan, usageChan, err := agentStrategy.Model.Chat(ctx, clientID, text)
	if err != nil {
		logger.Error(ctx, "llm service failed", map[string]any{
			"error":    err.Error(),
			"clientID": clientID,
			"text":     text,
		})
		return err
	}

	// merge the token text from replyChan to text buffer
	textBuffer := buffer.NewTextBuffer(replyChan)
	// generate audio by every sentence
	for sentence := range textBuffer.Output() {
		// tts
		audio, err := ttsClient.Synthesize(ctx, sentence)
		if err != nil {
			logger.Error(ctx, "tts service failed", map[string]any{
				"error":      err.Error(),
				"clientID":   clientID,
				"sentence":   sentence,
				"audio_size": len(audio),
			})
			continue
		}
		// publish the audio message to the MQTT topic
		err = Publish(ctx, clientID, audio)
		if err != nil {
			logger.Error(ctx, "publish audio failed", map[string]any{
				"error":    err.Error(),
				"clientID": clientID,
				"text":     sentence,
			})
			return err
		}

		fmt.Printf("sentence: %s\n", sentence)
	}

	// record usage
	go func() {
		bgCtx := context.Background()
		tid := tools.GetTraceID(ctx)
		if tid != "" {
			bgCtx = tools.WithTraceID(bgCtx, tid)
		}

		select {
		case u, ok := <-usageChan:
			if !ok {
				logger.Info(bgCtx, "llm usage closed", map[string]any{
					"clientID": clientID,
				})
				return
			}
			logger.Info(bgCtx, "llm usage", map[string]any{
				"clientID": clientID,
				"usage":    u,
			})

			// record into metering repository
			err = meteringService.Record(bgCtx, clientID, *u)
			if err != nil {
				logger.Error(bgCtx, "metering record failed", map[string]any{
					"error":    err.Error(),
					"clientID": clientID,
					"usage":    u,
				})
			}

		case <-time.After(10 * time.Second): // ⏱️ 超时兜底
			logger.Warn(bgCtx, "timeout waiting for LLM usage data", map[string]any{
				"clientID": clientID,
			})
		}
	}()

	return nil
}

func SpecialAction(ctx context.Context, clientID string, intent *nlu.Intent) error {
	if intent == nil {
		return fmt.Errorf("intent is nil")
	}

	switch intent.Text {
	case constant.IntentDenyAction: // cannel action -- the unique command
		// TODO publish the cancel message to the MQTT topic

		return nil
	case constant.IntentPlayMusic: // play music action
		// TODO publish the play music message to the MQTT topic

		return nil
	case constant.IntentSetTemperature: // set temperature action
		// TODO publish the set temperature message to the MQTT topic

		return nil
	}
	return nil
}

// Publish publish the message to the MQTT topic
// Parameters:
//   - ctx: the context.Context object
//   - clientID: the device sequence number to generate the MQTT topic
//   - payload: the message payload to publish
//
// Returns:
//   - error: the error object if the publish failed
func Publish(ctx context.Context, clientID string, payload []byte) error {
	// TODO 移除
	// @dev 暂存到./storage/tmp/audio/<clientID>/<timestamp>.wav
	rootDir := tools.GetRootDir()
	audioDir := filepath.Join(rootDir, "storage", "tmp", "audio", clientID)
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		logger.Error(ctx, "failed to create audio dir", map[string]any{
			"dir":   audioDir,
			"error": err.Error(),
		})
		return err
	}
	filename := fmt.Sprintf("%d_%04d.wav", time.Now().UnixNano(), atomic.AddUint64(&publishCounter, 1))
	fullPath := filepath.Join(audioDir, filename)

	if err := os.WriteFile(fullPath, payload, 0644); err != nil {
		logger.Error(ctx, "failed to save audio file", map[string]any{
			"path":  fullPath,
			"error": err.Error(),
		})
		return err
	}

	logger.Info(ctx, "audio saved", map[string]any{
		"path":       fullPath,
		"size_bytes": len(payload),
		"client_id":  clientID,
	})

	// TODO 发布MQTT消息
	// TODO 开发测试目前先保留配置参数硬编码，记得移除
	// test topic：test/T0001/A0001/voice/client
	topic := mqttCore.Topic{ // TODO: get from device registry
		Vendor:      constant.VendorTest,
		DeviceType:  "T0001",
		DeviceSN:    clientID,
		CommandType: "voice",
		Flag:        "client",
	}
	mqtt, err := mqttCore.GetMQTTClient(ctx, topic)
	if err != nil {
		logger.Error(ctx, "failed to get mqtt client", map[string]any{
			"clientID": clientID,
			"error":    err.Error(),
		})
		return err
	}

	audioConfig := voice.AudioConfig{
		AudioFormat:     mqttCommon.VoiceAudioFormatWav,
		AudioSampleRate: 16000, // TODO: get from constant
		AudioChannel:    1,
	}

	logger.Info(ctx, "publish audio config", map[string]any{
		"topic":        topic.String(),
		"audio_config": audioConfig,
	})

	err = mqtt.Publish(ctx, payload, audioConfig)
	if err != nil {
		logger.Error(ctx, "failed to publish audio stream", map[string]any{
			"clientID": clientID,
			"error":    err.Error(),
		})
		return err
	}

	return nil
}
