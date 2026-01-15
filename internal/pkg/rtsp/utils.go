package rtsp

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// GenerateSessionID generates a random session ID
func GenerateSessionID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	
	// Convert to hex string
	var s [32]byte
	for i, bb := range b {
		s[i*2] = "0123456789abcdef"[bb>>4]
		s[i*2+1] = "0123456789abcdef"[bb&0xf]
	}
	
	return string(s[:]), nil
}

// GenerateSequenceNumber generates a random sequence number
func GenerateSequenceNumber() (int, error) {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return 0, err
	}
	
	// Convert to positive integer
	num := (int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])) & 0x7FFFFFFF
	return num, nil
}

// ParseTransportHeader parses the Transport header value
func ParseTransportHeader(transportValue string) (map[string]string, error) {
	params := make(map[string]string)
	
	// Split by semicolon to get individual parameters
	parts := strings.Split(transportValue, ";")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		
		// Handle the transport protocol specification (e.g., RTP/AVP)
		if !strings.Contains(part, "=") {
			params["transport"] = part
			continue
		}
		
		// Split key=value pairs
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			
			// Remove quotes if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			
			params[key] = value
		}
	}
	
	return params, nil
}

// FormatTransportHeader formats a transport header from parameters
func FormatTransportHeader(params map[string]string) string {
	var parts []string
	
	// Add transport protocol first
	if transport, exists := params["transport"]; exists {
		parts = append(parts, transport)
		delete(params, "transport")
	} else {
		// Default to RTP/AVP
		parts = append(parts, TransportRTPAVP)
	}
	
	// Add unicast/multicast
	if delivery, exists := params["delivery"]; exists {
		parts = append(parts, delivery)
		delete(params, "delivery")
	} else {
		// Default to unicast
		parts = append(parts, TransportUnicast)
	}
	
	// Add remaining parameters
	for key, value := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	
	return strings.Join(parts, ";")
}

// IsSupportedMethod checks if the method is supported
func IsSupportedMethod(method string) bool {
	switch method {
	case Options, Describe, Setup, Play, Pause, Teardown:
		return true
	default:
		return false
	}
}

// GetSupportedMethods returns a list of supported methods
func GetSupportedMethods() []string {
	return []string{Options, Describe, Setup, Play, Pause, Teardown}
}

// FormatSupportedMethods returns supported methods as a comma-separated string
func FormatSupportedMethods() string {
	methods := GetSupportedMethods()
	return strings.Join(methods, ", ")
}

// GetStatusCodeText returns the text for a given status code
func GetStatusCodeText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 400:
		return "Bad Request"
	case 404:
		return "Not Found"
	case 453:
		return "Not Enough Bandwidth"
	case 454:
		return "Session Not Found"
	case 455:
		return "Method Not Valid in This State"
	case 456:
		return "Header Field Not Valid for Resource"
	case 457:
		return "Invalid Range"
	case 458:
		return "Parameter Not Understood"
	case 459:
		return "Conference Not Found"
	case 460:
		return "Bandwidth Parameter Not Understood"
	case 461:
		return "Unsupported Transport"
	case 500:
		return "Internal Server Error"
	case 501:
		return "Not Implemented"
	case 503:
		return "Service Unavailable"
	case 505:
		return "RTSP Version Not Supported"
	default:
		return "Unknown Status"
	}
}