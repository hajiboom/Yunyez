// Package format implements media format handling
// It parses H.264 codec configuration from SDP attributes
package format

import (
	"encoding/hex"
	"fmt"
)

// H264Config represents H.264 codec configuration
type H264Config struct {
	ProfileLevelID string // Profile level ID
	SPS            []byte // Sequence Parameter Set
	PPS            []byte // Picture Parameter Set
	SPSID          uint8  // SPS ID
	PPSID          uint8  // PPS ID
}

// ParseH264Config extracts H.264 configuration from SDP attributes
func ParseH264Config(attributes []string) (*H264Config, error) {
	config := &H264Config{}

	for _, attr := range attributes {
		if len(attr) > 10 && attr[:10] == "configuration=" {
			configStr := attr[12:] // Remove "configuration:" prefix and quotes
			if len(configStr) >= 2 {
				config.ProfileLevelID = configStr[:6] // First 6 chars are profile-level-id
				remaining := configStr[6:]

				// Decode SPS and PPS from hex string
				spspps, err := hex.DecodeString(remaining)
				if err != nil {
					return nil, fmt.Errorf("failed to decode SPS/PPS: %v", err)
				}

				// Parse SPS and PPS from the byte array
				err = parseSPSPPS(spspps, config)
				if err != nil {
					return nil, fmt.Errorf("failed to parse SPS/PPS: %v", err)
				}
			}
		}
	}

	return config, nil
}

// parseSPSPPS parses SPS and PPS from the byte array
func parseSPSPPS(data []byte, config *H264Config) error {
	if len(data) < 8 {
		return fmt.Errorf("data too short to contain SPS/PPS")
	}

	// First 6 bytes are typically profile-level-id and other parameters
	// Then we have SPS and PPS lengths
	pos := 6

	// Read SPS count
	spsCount := int(data[pos])
	pos++

	for i := 0; i < spsCount; i++ {
		if pos+2 > len(data) {
			return fmt.Errorf("not enough data for SPS length")
		}

		// Read SPS length
		spsLen := int(data[pos])<<8 | int(data[pos+1])
		pos += 2

		if pos+spsLen > len(data) {
			return fmt.Errorf("not enough data for SPS content")
		}

		// Extract SPS
		config.SPS = data[pos : pos+spsLen]
		pos += spsLen
	}

	// Read PPS count
	ppsCount := int(data[pos])
	pos++

	for i := 0; i < ppsCount; i++ {
		if pos+2 > len(data) {
			return fmt.Errorf("not enough data for PPS length")
		}

		// Read PPS length
		ppsLen := int(data[pos])<<8 | int(data[pos+1])
		pos += 2

		if pos+ppsLen > len(data) {
			return fmt.Errorf("not enough data for PPS content")
		}

		// Extract PPS
		config.PPS = data[pos : pos+ppsLen]
		pos += ppsLen
	}

	return nil
}

// GenerateSDPAttributes generates SDP attributes for H.264
func (c *H264Config) GenerateSDPAttributes() []string {
	var attrs []string

	// Add fmtp attribute for H.264
	if c.SPS != nil && c.PPS != nil {
		configStr := c.ProfileLevelID

		// Encode SPS and PPS
		var spspps []byte
		spspps = append(spspps, 0x01)                            // Version
		spspps = append(spspps, []byte(c.ProfileLevelID)[:3]...) // Profile level ID

		// Add SPS
		spspps = append(spspps, 0xff)                  // Reserved + SPS count (1)
		spspps = append(spspps, byte(len(c.SPS)>>8))   // SPS length high byte
		spspps = append(spspps, byte(len(c.SPS)&0xff)) // SPS length low byte
		spspps = append(spspps, c.SPS...)              // SPS data

		// Add PPS
		spspps = append(spspps, 0x01)                  // PPS count (1)
		spspps = append(spspps, byte(len(c.PPS)>>8))   // PPS length high byte
		spspps = append(spspps, byte(len(c.PPS)&0xff)) // PPS length low byte
		spspps = append(spspps, c.PPS...)              // PPS data

		// Convert to hex string
		hexStr := hex.EncodeToString(spspps)
		configStr += hexStr

		attrs = append(attrs, fmt.Sprintf(`fmtp:96 profile-level-id=%s; sprop-parameter-sets=%s; packetization-mode=1`,
			c.ProfileLevelID, configStr))
	}

	return attrs
}

// IsValid checks if the H264 config is valid
func (c *H264Config) IsValid() bool {
	return c.SPS != nil && c.PPS != nil && len(c.SPS) > 0 && len(c.PPS) > 0
}

// GetVideoParams returns video parameters from SPS
func (c *H264Config) GetVideoParams() (width, height, fps int, err error) {
	if !c.IsValid() {
		return 0, 0, 0, fmt.Errorf("invalid H264 config")
	}

	// Parse width, height, fps from SPS
	// This is a simplified implementation - a full implementation would require
	// a complete H.264 SPS parser
	width = 1920  // Placeholder
	height = 1080 // Placeholder
	fps = 30      // Placeholder

	return width, height, fps, nil
}
