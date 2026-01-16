// Package sdp provides functions for parsing and manipulating SDP session descriptions
// as defined in RFC 4566.
package sdp

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSessionDescription parses an SDP session description from a string
func ParseSessionDescription(sdpStr string) (*SessionDescription, error) {
	desc := &SessionDescription{}
	lines := strings.Split(sdpStr, "\n")

	var currentMedia *MediaDescription

	for i, line := range lines {
		// Remove carriage return and trim whitespace
		line = strings.TrimRight(line, "\r")
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse the line type and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid SDP line %d: %s", i+1, line)
		}

		lineType := parts[0]
		value := parts[1]

		switch lineType {
		case "v":
			version, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid version in SDP: %v", err)
			}
			desc.Version = version
		case "o":
			origin, err := parseOrigin(value)
			if err != nil {
				return nil, err
			}
			desc.Origin = *origin
		case "s":
			desc.SessionName = value
		case "i":
			desc.SessionInfo = value
		case "u":
			desc.URI = value
		case "e":
			desc.Emails = append(desc.Emails, value)
		case "p":
			desc.Phones = append(desc.Phones, value)
		case "c":
			connection, err := parseConnection(value)
			if err != nil {
				return nil, err
			}
			desc.Connection = connection
		case "b":
			bandwidth, err := parseBandwidth(value)
			if err != nil {
				return nil, err
			}
			desc.Bandwidth = append(desc.Bandwidth, *bandwidth)
		case "t":
			timing, err := parseTiming(value)
			if err != nil {
				return nil, err
			}
			desc.Timing = timing
		case "r":
			repeatTime, err := parseRepeatTime(value)
			if err != nil {
				return nil, err
			}
			desc.RepeatTimes = append(desc.RepeatTimes, *repeatTime)
		case "z":
			timeZone, err := parseTimeZone(value)
			if err != nil {
				return nil, err
			}
			desc.TimeZones = append(desc.TimeZones, *timeZone)
		case "k":
			desc.EncryptionKey = value
		case "a":
			attr, err := parseAttribute(value)
			if err != nil {
				return nil, err
			}
			if currentMedia != nil {
				currentMedia.Attributes = append(currentMedia.Attributes, *attr)
			} else {
				desc.Attributes = append(desc.Attributes, *attr)
			}
		case "m":
			mediaDesc, err := parseMediaDescription(value)
			if err != nil {
				return nil, err
			}
			desc.MediaDesc = append(desc.MediaDesc, mediaDesc)
			currentMedia = mediaDesc
		default:
			// Ignore unknown line types
		}
	}

	return desc, nil
}

// String returns the SDP string representation
func (desc *SessionDescription) String() string {
	var sdpLines []string

	// Version
	sdpLines = append(sdpLines, fmt.Sprintf("v=%d", desc.Version))

	// Origin
	sdpLines = append(sdpLines, fmt.Sprintf("o=%s", desc.Origin.String()))

	// Session name
	sdpLines = append(sdpLines, fmt.Sprintf("s=%s", desc.SessionName))

	// Session info (if present)
	if desc.SessionInfo != "" {
		sdpLines = append(sdpLines, fmt.Sprintf("i=%s", desc.SessionInfo))
	}

	// URI (if present)
	if desc.URI != "" {
		sdpLines = append(sdpLines, fmt.Sprintf("u=%s", desc.URI))
	}

	// Emails
	for _, email := range desc.Emails {
		sdpLines = append(sdpLines, fmt.Sprintf("e=%s", email))
	}

	// Phones
	for _, phone := range desc.Phones {
		sdpLines = append(sdpLines, fmt.Sprintf("p=%s", phone))
	}

	// Connection (if present)
	if desc.Connection != nil {
		sdpLines = append(sdpLines, fmt.Sprintf("c=%s", desc.Connection.String()))
	}

	// Bandwidth
	for _, bw := range desc.Bandwidth {
		sdpLines = append(sdpLines, fmt.Sprintf("b=%s", bw.String()))
	}

	// Timing
	if desc.Timing != nil {
		sdpLines = append(sdpLines, fmt.Sprintf("t=%s", desc.Timing.String()))
	}

	// Repeat times
	for _, rt := range desc.RepeatTimes {
		sdpLines = append(sdpLines, fmt.Sprintf("r=%s", rt.String()))
	}

	// Time zones
	for _, tz := range desc.TimeZones {
		sdpLines = append(sdpLines, fmt.Sprintf("z=%s", tz.String()))
	}

	// Encryption key (if present)
	if desc.EncryptionKey != "" {
		sdpLines = append(sdpLines, fmt.Sprintf("k=%s", desc.EncryptionKey))
	}

	// Session attributes
	for _, attr := range desc.Attributes {
		sdpLines = append(sdpLines, fmt.Sprintf("a=%s", attr.String()))
	}

	// Media descriptions
	for _, media := range desc.MediaDesc {
		sdpLines = append(sdpLines, fmt.Sprintf("m=%s", media.MediaName))

		// Media title (if present)
		if media.MediaTitle != "" {
			sdpLines = append(sdpLines, fmt.Sprintf("i=%s", media.MediaTitle))
		}

		// Media connection (if present and different from session connection)
		if media.Connection != nil && (desc.Connection == nil ||
			media.Connection.NetType != desc.Connection.NetType ||
			media.Connection.AddrType != desc.Connection.AddrType ||
			media.Connection.ConnectionAddr != desc.Connection.ConnectionAddr) {
			sdpLines = append(sdpLines, fmt.Sprintf("c=%s", media.Connection.String()))
		}

		// Media bandwidth
		for _, bw := range media.Bandwidth {
			sdpLines = append(sdpLines, fmt.Sprintf("b=%s", bw.String()))
		}

		// Media encryption key (if present)
		if media.EncryptionKey != "" {
			sdpLines = append(sdpLines, fmt.Sprintf("k=%s", media.EncryptionKey))
		}

		// Media attributes
		for _, attr := range media.Attributes {
			sdpLines = append(sdpLines, fmt.Sprintf("a=%s", attr.String()))
		}
	}

	return strings.Join(sdpLines, "\r\n") + "\r\n"
}

// parseOrigin parses an origin line
func parseOrigin(value string) (*Origin, error) {
	parts := strings.Fields(value)
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid origin line: %s", value)
	}

	return &Origin{
		Username:       parts[0],
		SessionID:      parts[1],
		SessionVersion: parts[2],
		NetType:        parts[3],
		AddrType:       parts[4],
		UnicastAddr:    parts[5],
	}, nil
}

// String returns the origin string representation
func (o *Origin) String() string {
	return fmt.Sprintf("%s %s %s %s %s %s",
		o.Username, o.SessionID, o.SessionVersion, o.NetType, o.AddrType, o.UnicastAddr)
}

// parseConnection parses a connection line
func parseConnection(value string) (*ConnectionData, error) {
	parts := strings.Fields(value)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid connection line: %s", value)
	}

	conn := &ConnectionData{
		NetType:        parts[0],
		AddrType:       parts[1],
		ConnectionAddr: parts[2],
	}

	// Parse optional TTL and number of addresses
	if len(parts) >= 4 {
		ttl, err := strconv.Atoi(parts[3])
		if err != nil {
			return nil, fmt.Errorf("invalid TTL in connection line: %v", err)
		}
		conn.TTL = ttl

		if len(parts) >= 5 {
			numAddr, err := strconv.Atoi(parts[4])
			if err != nil {
				return nil, fmt.Errorf("invalid number of addresses in connection line: %v", err)
			}
			conn.NumAddr = numAddr
		}
	}

	return conn, nil
}

// String returns the connection string representation
func (c *ConnectionData) String() string {
	if c.TTL > 0 {
		if c.NumAddr > 0 {
			return fmt.Sprintf("%s %s %s %d %d", c.NetType, c.AddrType, c.ConnectionAddr, c.TTL, c.NumAddr)
		}
		return fmt.Sprintf("%s %s %s %d", c.NetType, c.AddrType, c.ConnectionAddr, c.TTL)
	}
	return fmt.Sprintf("%s %s %s", c.NetType, c.AddrType, c.ConnectionAddr)
}

// parseBandwidth parses a bandwidth line
func parseBandwidth(value string) (*Bandwidth, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid bandwidth line: %s", value)
	}

	bandwidth, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid bandwidth value: %v", err)
	}

	return &Bandwidth{
		BandwidthType: parts[0],
		Bandwidth:     bandwidth,
	}, nil
}

// String returns the bandwidth string representation
func (b *Bandwidth) String() string {
	return fmt.Sprintf("%s:%d", b.BandwidthType, b.Bandwidth)
}

// parseTiming parses a timing line
func parseTiming(value string) (*Timing, error) {
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid timing line: %s", value)
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %v", err)
	}

	stop, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid stop time: %v", err)
	}

	return &Timing{
		Start: start,
		Stop:  stop,
	}, nil
}

// String returns the timing string representation
func (t *Timing) String() string {
	return fmt.Sprintf("%d %d", t.Start, t.Stop)
}

// parseRepeatTime parses a repeat time line
func parseRepeatTime(value string) (*RepeatTime, error) {
	parts := strings.Fields(value)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repeat time line: %s", value)
	}

	intervals, err := parseInt64Slice(parts[0:2])
	if err != nil {
		return nil, fmt.Errorf("invalid intervals in repeat time: %v", err)
	}

	var offsets []int64
	if len(parts) > 2 {
		offsets, err = parseInt64Slice(parts[2:])
		if err != nil {
			return nil, fmt.Errorf("invalid offsets in repeat time: %v", err)
		}
	}

	return &RepeatTime{
		Intervals: intervals,
		Offsets:   offsets,
	}, nil
}

// String returns the repeat time string representation
func (r *RepeatTime) String() string {
	var parts []string
	for _, interval := range r.Intervals {
		parts = append(parts, strconv.FormatInt(interval, 10))
	}
	for _, offset := range r.Offsets {
		parts = append(parts, strconv.FormatInt(offset, 10))
	}
	return strings.Join(parts, " ")
}

// parseTimeZone parses a time zone line
func parseTimeZone(value string) (*TimeZone, error) {
	parts := strings.Fields(value)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid time zone line: %s", value)
	}

	adjustmentTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid adjustment time: %v", err)
	}

	offset, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid offset: %v", err)
	}

	return &TimeZone{
		AdjustmentTime: adjustmentTime,
		Offset:         offset,
	}, nil
}

// String returns the time zone string representation
func (z *TimeZone) String() string {
	return fmt.Sprintf("%d %d", z.AdjustmentTime, z.Offset)
}

// parseAttribute parses an attribute line
func parseAttribute(value string) (*Attribute, error) {
	var key, val string

	if idx := strings.Index(value, ":"); idx != -1 {
		key = value[:idx]
		val = value[idx+1:]
	} else {
		key = value
	}

	return &Attribute{
		Key:   key,
		Value: val,
	}, nil
}

// String returns the attribute string representation
func (a *Attribute) String() string {
	if a.Value != "" {
		return fmt.Sprintf("%s:%s", a.Key, a.Value)
	}
	return a.Key
}

// parseMediaDescription parses a media description line
func parseMediaDescription(value string) (*MediaDescription, error) {
	parts := strings.Fields(value)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid media description line: %s", value)
	}

	return &MediaDescription{
		MediaName: value, // Keep the full value as media name
	}, nil
}

// parseInt64Slice converts a slice of strings to a slice of int64
func parseInt64Slice(strs []string) ([]int64, error) {
	result := make([]int64, len(strs))
	for i, str := range strs {
		num, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, err
		}
		result[i] = num
	}
	return result, nil
}
