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
	config "yunyez/internal/common/config"
	constant "yunyez/internal/common/constant"
	asr "yunyez/internal/pkg/agent/asr"
	llm "yunyez/internal/pkg/agent/llm"
	nlu "yunyez/internal/pkg/agent/nlu"
	tts "yunyez/internal/pkg/agent/tts"
	logger "yunyez/internal/pkg/logger"
	buffer "yunyez/internal/service/voice/buffer"
)

var (
	asrModel  = config.GetString("asr.model")   // default asr model
	nluModel  = config.GetString("nlu.model")   // default nlu model
	chatModel = config.GetString("agent.model") // default chat model
	ttsModel  = config.GetString("tts.model")   // default tts model

	asrEndpoint = config.GetString("asr.endpoint") // default asr endpoint

	agentStrategy *llm.Strategy // llm model strategy

	asrClient asr.Service
	nluClient nlu.Client
	ttsClient tts.Service
)

func init() {
	// 初始化智能体模型

	asrClient = asr.NewASRClient(asrModel, asrEndpoint) // init asr client

	agentStrategy = &llm.Strategy{}
	agentStrategy.SetAgent(chatModel)

	ttsClient = tts.NewTTSClient() // init tts client
}

// ChatPipeline response the natural language conversation response
// step:
// 1. asr: recognize the voice message to text
// 2. nlu: understand the text message to intent and entities
// 3. chat: call the llm model to chat
// 4. tts: generate the audio data from the chat response by tts
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
	replyChan, err := agentStrategy.Model.Chat(ctx, text)
	if err != nil {
		logger.Error(ctx, "llm service failed", map[string]any{
			"error":    err.Error(),
			"clientID": clientID,
			"text":     text,
		})
		return err
	}
	// merge the replyChan to text buffer
	textBuffer := buffer.NewTextBuffer(replyChan)
	// generate audio by every sentence
	for sentence := range textBuffer.Output() {
		// tts
		audio, err := ttsClient.Synthesize(ctx, sentence)
		if err != nil {
			logger.Error(ctx, "tts service failed", map[string]any{
				"error":    err.Error(),
				"clientID": clientID,
				"sentence": sentence,
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
	}

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
	// TODO 发布MQTT消息
	return nil
}
