package rtsp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

// Parser handles parsing of RTSP messages
type Parser struct {
	reader *textproto.Reader
}

// NewParser creates a new RTSP parser
func NewParser(r io.Reader) *Parser {
	br := bufio.NewReader(r)
	tp := textproto.NewReader(br)
	
	return &Parser{
		reader: tp,
	}
}

// ParseRequest parses an RTSP request from the reader
func (p *Parser) ParseRequest() (*Request, error) {
	// Read the request line
	requestLine, err := p.reader.ReadLine()
	if err != nil {
		return nil, err
	}

	// Parse the request line
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
	for {
		line, err := p.reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if line == "" {
			// End of headers
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

	// Get content length to determine if there's a body
	contentLength := 0
	if clStr, exists := headers[ContentLengthHeader]; exists {
		cl, err := strconv.Atoi(clStr)
		if err == nil {
			contentLength = cl
		}
	}

	// Read body if present
	var body []byte
	if contentLength > 0 {
		body = make([]byte, contentLength)
		_, err := io.ReadFull(p.reader.R, body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %v", err)
		}
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

	// Construct the raw data for the request
	var rawBuf bytes.Buffer
	rawBuf.WriteString(requestLine + "\r\n")
	for k, v := range headers {
		rawBuf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	rawBuf.WriteString("\r\n") // Empty line separating headers and body
	if len(body) > 0 {
		rawBuf.Write(body)
	}

	return &Request{
		Method:    method,
		URL:       parsedURL,
		Version:   version,
		Headers:   headers,
		Body:      body,
		Raw:       rawBuf.Bytes(),
		CSeq:      cseq,
		SessionID: sessionID,
	}, nil
}

// ParseResponse parses an RTSP response from the reader
func (p *Parser) ParseResponse() (*Response, error) {
	// Read the status line
	statusLine, err := p.reader.ReadLine()
	if err != nil {
		return nil, err
	}

	// Parse the status line
	parts := strings.Fields(statusLine)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid RTSP status line: %s", statusLine)
	}

	version := parts[0]
	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", parts[1])
	}
	
	// Status text is everything after the status code
	statusText := strings.Join(parts[2:], " ")

	// Parse headers
	headers := make(map[string]string)
	for {
		line, err := p.reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if line == "" {
			// End of headers
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

	// Get content length to determine if there's a body
	contentLength := 0
	if clStr, exists := headers[ContentLengthHeader]; exists {
		cl, err := strconv.Atoi(clStr)
		if err == nil {
			contentLength = cl
		}
	}

	// Read body if present
	var body []byte
	if contentLength > 0 {
		body = make([]byte, contentLength)
		_, err := io.ReadFull(p.reader.R, body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %v", err)
		}
	}

	// Construct the raw data for the response
	var rawBuf bytes.Buffer
	rawBuf.WriteString(statusLine + "\r\n")
	for k, v := range headers {
		rawBuf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	rawBuf.WriteString("\r\n") // Empty line separating headers and body
	if len(body) > 0 {
		rawBuf.Write(body)
	}

	return &Response{
		Version:    version,
		StatusCode: statusCode,
		StatusText: statusText,
		Headers:    headers,
		Body:       body,
		Raw:        rawBuf.Bytes(),
	}, nil
}

// ParseRequestFromBytes parses an RTSP request from a byte slice
func ParseRequestFromBytes(data []byte) (*Request, error) {
	reader := bytes.NewReader(data)
	parser := NewParser(reader)
	
	req, err := parser.ParseRequest()
	if err != nil {
		return nil, err
	}
	
	// Since we're parsing from bytes, we need to reconstruct the raw data properly
	req.Raw = data
	
	return req, nil
}

// ParseResponseFromBytes parses an RTSP response from a byte slice
func ParseResponseFromBytes(data []byte) (*Response, error) {
	reader := bytes.NewReader(data)
	parser := NewParser(reader)
	
	resp, err := parser.ParseResponse()
	if err != nil {
		return nil, err
	}
	
	// Since we're parsing from bytes, we need to reconstruct the raw data properly
	resp.Raw = data
	
	return resp, nil
}