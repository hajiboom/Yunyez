// Package voicemanager 语音管理
package voicemanager

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"yunyez/internal/pkg/logger"
	mqtt_common "yunyez/internal/pkg/mqtt/common"
	voice_fragment "yunyez/internal/service/voice/fragment"
	mqtt_voice "yunyez/internal/pkg/mqtt/protocol/voice"
)

//
// UploadVoice 语音处理
// @Summary 语音处理
// @Description 接收来自MQTT转发的语音消息
// @Tags 语音管理
// @Accept multipart/form-data
// @Produce json
// @Success 200 {object} gin.H{"message": "语音处理成功"}
// @Failure 400 {object} gin.H{"error": "读取请求体失败"}
// @Router /voice/upload [post]
func UploadVoice(c *gin.Context) {
	fmt.Println("UploadVoice start>>>>>")

	// 读取请求字节
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}
	// 解析消息音频头/音频载荷
	var header mqtt_voice.Header // 音频协议头 固定12字节
	// 解析音频协议头
	if err := header.UnmarshalHeader(body[:mqtt_voice.HeaderSize]); err != nil {
		logger.Error(c.Request.Context(), "mqtt.voice.Header.UnmarshalHeader failed", map[string]interface{}{
			"error": err.Error(),
			"topic": c.GetHeader("Topic"),
			"payload_len": len(body[mqtt_voice.HeaderSize:]),
			"ClientID": c.GetHeader("ClientID"),
		})
		return
	}
	// 音频数据处理
	switch header.F {
	case mqtt_common.VoiceFrameFull: // 完整帧
		if err := voice_fragment.ProcessFull(c.Request.Context(), c.GetHeader("ClientID"), &header, body[mqtt_voice.HeaderSize:]); err != nil {
			logger.Error(c.Request.Context(), "voice_fragment.ProcessFull failed", map[string]any{
				"error": err.Error(),
				"topic": c.GetHeader("Topic"),
				"payload_len": len(body[mqtt_voice.HeaderSize:]),
				"ClientID": c.GetHeader("ClientID"),
			})
			return
		}

		
	case mqtt_common.VoiceFrameFragment, mqtt_common.VoiceFrameLast: // 分片帧, 最后一帧
		if err := voice_fragment.ProcessFragment(c.Request.Context(), c.GetHeader("ClientID"), &header, body[mqtt_voice.HeaderSize:]); err != nil {
			logger.Error(c.Request.Context(), "voice_fragment.ProcessFragment failed", map[string]any{
				"error": err.Error(),
				"topic": c.GetHeader("Topic"),
				"payload_len": len(body[mqtt_voice.HeaderSize:]),
				"ClientID": c.GetHeader("ClientID"),
			})
			return
		}
	default:
		logger.Error(c.Request.Context(), "voiceManage.UploadVoice unknown frame type", map[string]any{
			"topic": c.GetHeader("Topic"),
			"payload_len": len(body[mqtt_voice.HeaderSize:]),
			"ClientID": c.GetHeader("ClientID"),
		})
		return
	}
}