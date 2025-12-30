// Package voice defines the binary protocol for voice data over MQTT.
/*
This package defines the binary protocol for voice data over MQTT.
It is NOT a general-purpose audio format, but specific to Yunyez MQTT communication.
If other transports (e.g., WebSocket) adopt this format, consider moving to pkg/protocol/.
*/
package voice

import (
	"encoding/binary"
	"errors"
)

const (
	HeaderSize = 12 // 96 bits = 12 bytes
)

// Header 音频协议头
type Header struct {
	Version     uint8  // 4 bits (0-15)
	AudioFormat uint8  // 8 bits
	SampleRate  uint16 // 16 bits
	Ch          uint8  // 2 bits (1=mono, 2=stereo, 3=multi)
	F           uint8  // 2 bits (1=full, 2=fragment, 3=last)
	FrameSeq    uint16 // 16 bits
	Timestamp   uint16 // 16 bits
	PayloadLen  uint16 // 16 bits
	CRC16       uint16 // 16 bits, computed over header (without CRC) + payload
}

// Marshal 序列化音频协议头
// @return []byte 序列化后的音频协议头数据
func (h *Header) Marshal() []byte {
	buf := make([]byte, HeaderSize)

	// Byte 0: [Ver:4][AUdioFormat high 4]
	buf[0] = ((h.Version & 0x0F) << 4) | ((h.AudioFormat >> 4) & 0x0F)

	// Byte 1: [AudioFormat low 4][SampleRate high 4]
	buf[1] = byte((h.AudioFormat&0x0F)<<4) | byte((h.SampleRate>>12)&0x0F)

	// Byte 2: [SampleRate mid 8]
	buf[2] = byte((h.SampleRate >> 4) & 0xFF)

	// Byte 3: [SampleRate low 4][Ch:2][F:2]
	buf[3] = byte((h.SampleRate&0x0F)<<4) | byte((h.Ch&0x03)<<2) | byte(h.F&0x03)

	// Bytes 4-5: FrameSeq
	binary.BigEndian.PutUint16(buf[4:6], h.FrameSeq)

	// Bytes 6-7: Timestamp
	binary.BigEndian.PutUint16(buf[6:8], h.Timestamp)

	// Bytes 8-9: PayloadLen
	binary.BigEndian.PutUint16(buf[8:10], h.PayloadLen)

	// Bytes 10-11: CRC16
	binary.BigEndian.PutUint16(buf[10:12], h.CRC16)

	return buf
}

// UnmarshalHeader 反序列化音频协议头
// @param data 音频协议头数据
// @return error 反序列化错误
func (h *Header) UnmarshalHeader(data []byte) error {
	if len(data) < HeaderSize {
		return errors.New("insufficient data for header (need 12 bytes)")
	}

	h.Version = (data[0] >> 4) & 0x0F
	h.AudioFormat = ((data[0] & 0x0F) << 4) | (data[1] >> 4)

	// Build SampleRate from 3 parts
	h.SampleRate = (uint16(data[1]&0x0F) << 12) |
		(uint16(data[2]) << 4) |
		(uint16(data[3]>>4) & 0x0F)

	h.Ch = (data[3] >> 2) & 0x03
	h.F = data[3] & 0x03
	h.FrameSeq = binary.BigEndian.Uint16(data[4:6])
	h.Timestamp = binary.BigEndian.Uint16(data[6:8])
	h.PayloadLen = binary.BigEndian.Uint16(data[8:10])
	h.CRC16 = binary.BigEndian.Uint16(data[10:12])

	return nil
}
