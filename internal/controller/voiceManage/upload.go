// Package voicemanager ai-chat audio http interface
package voicemanager

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"yunyez/internal/pkg/logger"
	mqttCommon "yunyez/internal/pkg/mqtt/common"
	voiceHandler "yunyez/internal/service/voice/handler"
	mqttVoice "yunyez/internal/pkg/mqtt/protocol/voice"
)

//
// UploadVoice audio upload 
// @Summary 语音处理
// @Description 接收来自MQTT转发的语音消息
// @Tags 语音管理
// @Accept multipart/form-data
// @Produce json
// @Success 200 {object} gin.H{"message": "语音处理成功"}
// @Failure 400 {object} gin.H{"error": "读取请求体失败"}
// @Router /voice/upload [post]
func UploadVoice(c *gin.Context) {
	// 读取请求字节
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}
	// 解析消息音频头/音频载荷
	var header mqttVoice.Header // 音频协议头 固定12字节
	// 解析音频协议头
	if err := header.UnmarshalHeader(body[:mqttVoice.HeaderSize]); err != nil {
		logger.Error(c.Request.Context(), "mqtt.voice.Header.UnmarshalHeader failed", map[string]interface{}{
			"error": err.Error(),
			"topic": c.GetHeader("Topic"),
			"payload_len": len(body[mqttVoice.HeaderSize:]),
			"ClientID": c.GetHeader("ClientID"),
		})
		return
	}
	// 音频数据处理
	switch header.F {
	case mqttCommon.VoiceFrameFull: // 完整帧
		if err := voiceHandler.ProcessFull(c.Request.Context(), c.GetHeader("ClientID"), &header, body[mqttVoice.HeaderSize:]); err != nil {
			logger.Error(c.Request.Context(), "voiceHandler.ProcessFull failed", map[string]any{
				"error": err.Error(),
				"topic": c.GetHeader("Topic"),
				"payload_len": len(body[mqttVoice.HeaderSize:]),
				"ClientID": c.GetHeader("ClientID"),
			})
			return
		}
	case mqttCommon.VoiceFrameFragment, mqttCommon.VoiceFrameLast: // 分片帧, 最后一帧
		if err := voiceHandler.ProcessFragment(c.Request.Context(), c.GetHeader("ClientID"), &header, body[mqttVoice.HeaderSize:]); err != nil {
			logger.Error(c.Request.Context(), "voiceHandler.ProcessFragment failed", map[string]any{
				"error": err.Error(),
				"topic": c.GetHeader("Topic"),
				"payload_len": len(body[mqttVoice.HeaderSize:]),
				"ClientID": c.GetHeader("ClientID"),
			})
			return
		}
	default:
		logger.Error(c.Request.Context(), "voiceManage.UploadVoice unknown frame type", map[string]any{
			"topic": c.GetHeader("Topic"),
			"payload_len": len(body[mqttVoice.HeaderSize:]),
			"ClientID": c.GetHeader("ClientID"),
		})
		return
	}
}