package video

import (
	"fmt"
	"time"

	"yunyez/internal/pkg/media/sdp"
	"yunyez/internal/video/types"
)

// SDPGenerator generates SDP descriptions for streams
type SDPGenerator struct{}

// NewSDPGenerator creates a new SDP generator
func NewSDPGenerator() *SDPGenerator {
	return &SDPGenerator{}
}

// GenerateSDPForStream generates an SDP description for a stream
func (sg *SDPGenerator) GenerateSDPForStream(stream *types.Stream) (*sdp.SessionDescription, error) {
	origin := sdp.Origin{
		Username:       "-",
		SessionID:      fmt.Sprintf("%d", time.Now().Unix()),
		SessionVersion: "1",
		NetType:        "IN",
		AddrType:       "IP4",
		UnicastAddr:    "127.0.0.1", // This should be the actual server address
	}

	sessionDesc := &sdp.SessionDescription{
		Version:     0,
		Origin:      origin,
		SessionName: stream.Name,
		SessionInfo: stream.Name + " Stream",
		Connection: &sdp.ConnectionData{
			NetType:        "IN",
			AddrType:       "IP4",
			ConnectionAddr: "127.0.0.1", // This should be the actual server address
		},
		Timing: &sdp.Timing{
			Start: 0, // 0 means session is not bounded by time
			Stop:  0,
		},
	}

	// Add media description based on stream format
	mediaDesc, err := sg.createMediaDescription(stream)
	if err != nil {
		return nil, err
	}

	sessionDesc.MediaDesc = append(sessionDesc.MediaDesc, mediaDesc)

	return sessionDesc, nil
}

// createMediaDescription creates a media description based on the stream format
func (sg *SDPGenerator) createMediaDescription(stream *types.Stream) (*sdp.MediaDescription, error) {
	var mediaDesc *sdp.MediaDescription

	switch stream.MediaFormat {
	case types.H264MediaFormat:
		mediaDesc = &sdp.MediaDescription{
			MediaName: fmt.Sprintf("video 0 RTP/AVP %d", 96), // 96 is dynamic payload type for H264
			Attributes: []sdp.Attribute{
				{Key: "rtpmap", Value: "96 H264/90000"}, // 90000 Hz for video
				{Key: "fmtp", Value: "96 profile-level-id=42e01f; packetization-mode=1"},
				{Key: "control", Value: "track1"},
			},
		}
	case types.H265MediaFormat:
		mediaDesc = &sdp.MediaDescription{
			MediaName: fmt.Sprintf("video 0 RTP/AVP %d", 98), // 98 is dynamic payload type for H265
			Attributes: []sdp.Attribute{
				{Key: "rtpmap", Value: "98 H265/90000"}, // 90000 Hz for video
				{Key: "fmtp", Value: "98 profile-space=0; profile-id=1; tier-flag=0; level-id=93"},
				{Key: "control", Value: "track1"},
			},
		}
	case types.PCMAMediaFormat:
		mediaDesc = &sdp.MediaDescription{
			MediaName: fmt.Sprintf("audio 0 RTP/AVP %d", 8), // 8 is payload type for PCMA
			Attributes: []sdp.Attribute{
				{Key: "rtpmap", Value: "8 PCMA/8000"}, // 8000 Hz for audio
				{Key: "control", Value: "track2"},
			},
		}
	case types.PCMUFormat:
		mediaDesc = &sdp.MediaDescription{
			MediaName: fmt.Sprintf("audio 0 RTP/AVP %d", 0), // 0 is payload type for PCMU
			Attributes: []sdp.Attribute{
				{Key: "rtpmap", Value: "0 PCMU/8000"}, // 8000 Hz for audio
				{Key: "control", Value: "track2"},
			},
		}
	default:
		// Default to H264
		mediaDesc = &sdp.MediaDescription{
			MediaName: fmt.Sprintf("video 0 RTP/AVP %d", 96),
			Attributes: []sdp.Attribute{
				{Key: "rtpmap", Value: "96 H264/90000"},
				{Key: "fmtp", Value: "96 profile-level-id=42e01f; packetization-mode=1"},
				{Key: "control", Value: "track1"},
			},
		}
	}

	return mediaDesc, nil
}

// GenerateSDPString generates an SDP string for a stream
func (sg *SDPGenerator) GenerateSDPString(stream *types.Stream) (string, error) {
	sdpDesc, err := sg.GenerateSDPForStream(stream)
	if err != nil {
		return "", err
	}

	return sdpDesc.String(), nil
}
