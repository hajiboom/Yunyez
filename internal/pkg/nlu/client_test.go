package nlu

import (
	"testing"
	"yunyez/internal/common/constant"
)

func TestPredictTurnOnLight(t *testing.T) {
	client := NewClient()
	
	// 测试用例1: 打开所有灯
	input := &Input{
		Text: "打开所有灯",
	}
	intent, err := client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_TurnOnLight {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_TurnOnLight, intent.Intent)
	}

	// 测试用例2: 打开卧室的灯
	input = &Input{
		Text: "打开卧室的灯",
	}
	intent, err = client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_TurnOnLight {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_TurnOnLight, intent.Intent)
	}
}

func TestPredictTurnOffLight(t *testing.T) {
	client := NewClient()
	
	// 测试用例1: 关闭所有灯
	input := &Input{
		Text: "关闭所有灯",
	}
	intent, err := client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_TurnOffLight {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_TurnOffLight, intent.Intent)
	}

	// 测试用例2: 关闭客厅的灯
	input = &Input{
		Text: "关闭客厅的灯",
	}
	intent, err = client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_TurnOffLight {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_TurnOffLight, intent.Intent)
	}
}

func TestPredictSetTemperature(t *testing.T) {
	client := NewClient()
	
	// 测试用例1: 设置温度为25摄氏度
	input := &Input{
		Text: "设置温度为25摄氏度",
	}
	intent, err := client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_SetTemperature {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_SetTemperature, intent.Intent)
	}

	// 测试用例2: 空调温度调到26度
	input = &Input{
		Text: "空调温度调到26度",
	}
	intent, err = client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_SetTemperature {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_SetTemperature, intent.Intent)
	}
}

func TestPredictPlayMusic(t *testing.T) {
	client := NewClient()
	
	// 测试用例1: 我想听音乐
	input := &Input{
		Text: "我想听音乐",
	}
	intent, err := client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_PlayMusic {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_PlayMusic, intent.Intent)
	}

	// 测试用例2: 播放周杰伦的歌
	input = &Input{
		Text: "播放周杰伦的歌",
	}
	intent, err = client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_PlayMusic {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_PlayMusic, intent.Intent)
	}
}

func TestPredictChitChat(t *testing.T) {
	client := NewClient()
	
	// 测试用例1: 你好啊
	input := &Input{
		Text: "你好啊",
	}
	intent, err := client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_ChitChat {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_ChitChat, intent.Intent)
	}

	// 测试用例2: 早上好
	input = &Input{
		Text: "早上好",
	}
	intent, err = client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_ChitChat {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_ChitChat, intent.Intent)
	}
}

func TestPredictDenyAction(t *testing.T) {
	client := NewClient()
	
	// 测试用例1: 不需要开空调
	input := &Input{
		Text: "不需要开空调",
	}
	intent, err := client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_DenyAction {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_DenyAction, intent.Intent)
	}

	// 测试用例2: 不用关灯
	input = &Input{
		Text: "不用关灯",
	}
	intent, err = client.Predict(input)
	if err != nil {
		t.Errorf("Predict failed: %v", err)
	}
	if intent.Intent != constant.Intent_DenyAction {
		t.Errorf("Predict failed: expect %s, got %s", constant.Intent_DenyAction, intent.Intent)
	}
}
