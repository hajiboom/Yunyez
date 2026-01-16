package rtsp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Request represents an RTSP request
type Request struct {
	Method      string
	URL         *url.URL
	Version     string
	Headers     map[string]string
	Body        []byte
	Raw         []byte
	CSeq        int // Command sequence number
	SessionID   string
}

// Response represents an RTSP response
type Response struct {
	Version     string
	StatusCode  int
	StatusText  string
	Headers     map[string]string
	Body        []byte
	Raw         []byte
}

// ParseRequest parses an RTSP request from raw bytes
func ParseRequest(data []byte) (*Request, error) {
	lines := strings.Split(string(data), "\r\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("invalid RTSP request: empty data")
	}

	// Parse request line
	requestLine := lines[0]
	parts := strings.Fields(requestLine)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid RTSP request line: %s", requestLine)
	}

	method := parts[0]
	urlStr := parts[1]
	version := parts[2]

	// Validate RTSP version
	if version != RTSPVersion {
		return nil, fmt.Errorf("unsupported RTSP version: %s", version)
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Parse headers
	headers := make(map[string]string)
	var body []byte
	
	headerStart := 1
	for i := headerStart; i < len(lines); i++ {
		line := lines[i]
		
		// Empty line indicates end of headers
		if line == "" {
			// Everything after the empty line is the body
			if i+1 < len(lines) {
				bodyStr := strings.Join(lines[i+1:], "\r\n")
				body = []byte(bodyStr)
			}
			break
		}
		
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue // Skip malformed header
		}
		
		key := strings.TrimSpace(line[:colonIndex])
		value := strings.TrimSpace(line[colonIndex+1:])
		headers[key] = value
	}

	// Extract CSeq
	cseqStr, exists := headers[CSeqHeader]
	if !exists {
		return nil, fmt.Errorf("missing CSeq header")
	}
	
	cseq, err := strconv.Atoi(cseqStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CSeq value: %s", cseqStr)
	}

	// Extract session ID if present
	sessionID := headers[SessionHeader]

	return &Request{
		Method:    method,
		URL:       parsedURL,
		Version:   version,
		Headers:   headers,
		Body:      body,
		Raw:       data,
		CSeq:      cseq,
		SessionID: sessionID,
	}, nil
}

// String returns the string representation of the request
func (req *Request) String() string {
	requestLine := fmt.Sprintf("%s %s %s\r\n", req.Method, req.URL.String(), req.Version)
	
	var headers string
	for key, value := range req.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	
	result := requestLine + headers + "\r\n"
	if len(req.Body) > 0 {
		result += string(req.Body)
	}
	
	return result
}

// Bytes returns the byte representation of the request
func (req *Request) Bytes() []byte {
	return []byte(req.String())
}

// String returns the string representation of the response
func (resp *Response) String() string {
	statusLine := fmt.Sprintf("%s %d %s\r\n", resp.Version, resp.StatusCode, resp.StatusText)
	
	var headers string
	for key, value := range resp.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	
	result := statusLine + headers + "\r\n"
	if len(resp.Body) > 0 {
		result += string(resp.Body)
	}
	
	return result
}

// Bytes returns the byte representation of the response
func (resp *Response) Bytes() []byte {
	return []byte(resp.String())
}

// NewResponse creates a new RTSP response
func NewResponse(statusCode int, statusText string, headers map[string]string, body []byte) *Response {
	if headers == nil {
		headers = make(map[string]string)
	}
	
	return &Response{
		Version:    RTSPVersion,
		StatusCode: statusCode,
		StatusText: statusText,
		Headers:    headers,
		Body:       body,
	}
}

// AddHeader adds a header to the response
func (resp *Response) AddHeader(key, value string) {
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}
	resp.Headers[key] = value
}

// SetBody sets the response body and updates Content-Length header
func (resp *Response) SetBody(body []byte) {
	resp.Body = body
	resp.AddHeader(ContentLengthHeader, strconv.Itoa(len(body)))
}