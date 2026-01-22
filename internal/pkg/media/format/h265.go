package format

import (
	"encoding/hex"
	"fmt"
)

// H265Config represents H.265/HEVC codec configuration
type H265Config struct {
	ProfileSpace   string // Profile space
	ProfileID      string // Profile ID
	ProfileTier    string // Profile tier
	LevelID        string // Level ID
	Interlaced     bool   // Interlaced flag
	MaxBitRate     int    // Maximum bitrate
	MaxWidth       int    // Maximum width
	MaxHeight      int    // Maximum height
	ConstChromaID  bool   // Constant chroma format indicator
	ConstChromaFmt string // Constant chroma format
	ChromaFormat   string // Chroma format
	BitDepthLuma   string // Bit depth luma
	BitDepthChroma string // Bit depth chroma
	AvgBitRate     int    // Average bitrate
	DecPicBuffSize int    // Decoded picture buffer size
	SPS            []byte // Sequence Parameter Set
	PPS            []byte // Picture Parameter Set
	VPS            []byte // Video Parameter Set
}

// ParseH265Config extracts H.265 configuration from SDP attributes
func ParseH265Config(attributes []string) (*H265Config, error) {
	config := &H265Config{}

	for _, attr := range attributes {
		if len(attr) > 10 && attr[:10] == "sprop-vps=" {
			vpsStr := attr[10:] // Remove "sprop-vps=" prefix
			vps, err := hex.DecodeString(vpsStr)
			if err != nil {
				return nil, fmt.Errorf("failed to decode VPS: %v", err)
			}
			config.VPS = vps
		}

		if len(attr) > 10 && attr[:10] == "sprop-sps=" {
			spsStr := attr[10:] // Remove "sprop-sps=" prefix
			sps, err := hex.DecodeString(spsStr)
			if err != nil {
				return nil, fmt.Errorf("failed to decode SPS: %v", err)
			}
			config.SPS = sps
		}

		if len(attr) > 10 && attr[:10] == "sprop-pps=" {
			ppsStr := attr[10:] // Remove "sprop-pps=" prefix
			pps, err := hex.DecodeString(ppsStr)
			if err != nil {
				return nil, fmt.Errorf("failed to decode PPS: %v", err)
			}
			config.PPS = pps
		}

		if len(attr) > 10 && attr[:9] == "profile=" {
			profileStr := attr[9:] // Remove "profile=" prefix
			// Parse profile information
			config.ProfileID = profileStr
		}
	}

	return config, nil
}

// GenerateSDPAttributes generates SDP attributes for H.265
func (c *H265Config) GenerateSDPAttributes() []string {
	var attrs []string

	// Add fmtp attribute for H.265
	if c.SPS != nil && c.PPS != nil && c.VPS != nil {
		vpsHex := hex.EncodeToString(c.VPS)
		spsHex := hex.EncodeToString(c.SPS)
		ppsHex := hex.EncodeToString(c.PPS)

		attrs = append(attrs, fmt.Sprintf(
			`fmtp:98 profile-space=%s; profile-id=%s; tier-flag=%s; level-id=%s;interop-constraints=%s;sprop-vps=%s;sprop-sps=%s;sprop-pps=%s`,
			c.ProfileSpace, c.ProfileID, c.ProfileTier, c.LevelID, "B00000000000", vpsHex, spsHex, ppsHex))
	}

	return attrs
}

// IsValid checks if the H265 config is valid
func (c *H265Config) IsValid() bool {
	return c.SPS != nil && c.PPS != nil && c.VPS != nil &&
		len(c.SPS) > 0 && len(c.PPS) > 0 && len(c.VPS) > 0
}

// GetVideoParams returns video parameters from SPS
func (c *H265Config) GetVideoParams() (width, height, fps int, err error) {
	if !c.IsValid() {
		return 0, 0, 0, fmt.Errorf("invalid H265 config")
	}

	// Parse width, height, fps from SPS
	// This is a simplified implementation - a full implementation would require
	// a complete H.265 SPS parser
	width = 3840  // Placeholder
	height = 2160 // Placeholder
	fps = 30      // Placeholder

	return width, height, fps, nil
}