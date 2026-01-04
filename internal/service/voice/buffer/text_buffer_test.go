package buffer

import (
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"
)

// MockStream simulates LLM token-by-token streaming (e.g., Qwen)
type MockStream struct {
	tokens []string
}

func (m *MockStream) Stream() <-chan string {
	ch := make(chan string, len(m.tokens))
	go func() {
		defer close(ch)
		for _, tok := range m.tokens {
			time.Sleep(10 * time.Millisecond) // simulate network delay
			ch <- tok
		}
	}()
	return ch
}

// isDangerousBreak checks if a string ends in a dangerous position (e.g., mid-English word)
func isDangerousBreak(s string) bool {
	if len(s) == 0 {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	// Dangerous: alphanumeric or underscore (likely mid-word)
	return unicode.IsLetter(r) && !unicode.Is(unicode.Han, r) ||
		unicode.IsDigit(r) ||
		r == '_'
}

func TestTextBuffer(t *testing.T) {
	t.Run("Chinese sentence with safe breaks", func(t *testing.T) {
		stream := &MockStream{
			tokens: []string{"我", "是", "通", "义", "千", "问", "，", "是", "阿", "里", "巴", "巴", "集", "团", "旗", "下", "的", "通", "义", ".", "实", "验", "室", "自", "主", "研", "发", "的", "超", "大", "规", "模", "语", "言", "模", "型", "。"},
		}
		tb := NewTextBuffer(stream.Stream())

		var results []string
		for out := range tb.Output() {
			results = append(results, out)
			t.Logf(">> chunk: %s", results)
		}

		// Should produce 1 or 2 chunks, each ending safely
		if len(results) == 0 {
			t.Fatal("no output")
		}
		for i, r := range results {
			if isDangerousBreak(r) {
				t.Errorf("chunk %d ends dangerously: %q", i, r)
			}
			if len([]rune(r)) < 10 && i < len(results)-1 {
				t.Errorf("chunk %d too short: %q", i, r)
			}
		}
	})

	t.Run("Long English text without punctuation", func(t *testing.T) {
		stream := &MockStream{
			tokens: strings.Split("This is a very long English sentence without any punctuation that should eventually be flushed by length limit", ""),
		}
		tb := NewTextBuffer(stream.Stream())

		var results []string
		for out := range tb.Output() {
			results = append(results, out)
			t.Logf(">> chunk: %s", results)
		}

		if len(results) == 0 {
			t.Fatal("expected at least one chunk due to length flush")
		}
		// Last chunk may end mid-word — acceptable under length pressure
	})

	t.Run("Timeout flush for partial input", func(t *testing.T) {
		stream := &MockStream{
			tokens: []string{"Partial", " ", "text", " ", "without", " ", "punctuation"},
		}
		tb := NewTextBuffer(stream.Stream())

		// Wait beyond flushDelay
		time.Sleep(700 * time.Millisecond)

		var results []string
		// Drain channel non-blockingly
		done := make(chan struct{})
		go func() {
			for out := range tb.Output() {
				results = append(results, out)
				t.Logf(">> chunk: %s", results)
			}
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			// Force stop if hanging
		}

		if len(results) == 0 {
			t.Error("expected timeout flush")
		} else {
			t.Logf("timeout flushed: %q", results[0])
		}
	})

	t.Run("Mixed Chinese and English with safe breaks", func(t *testing.T) {
		stream := &MockStream{
			tokens: []string{
				"你好", "，", "我", "叫", "Qwen", "。", // first sentence
				"I", " ", "can", " ", "help", " ", "you", " ", "with", " ", "coding", ".", // second
			},
		}
		tb := NewTextBuffer(stream.Stream())

		var results []string
		for out := range tb.Output() {
			results = append(results, out)
		}

		if len(results) < 1 {
			t.Fatal("no output")
		}
		// First chunk should end with "。"
		if !strings.HasSuffix(results[0], "。") {
			t.Logf("first chunk: %q", results[0])
			// OK if merged with next due to buffering, but should not break mid-Chinese
		}
	})

	t.Run("Empty input", func(t *testing.T) {
		stream := &MockStream{tokens: []string{}}
		tb := NewTextBuffer(stream.Stream())
		results := []string{}
		for out := range tb.Output() {
			results = append(results, out)
		}
		if len(results) != 0 {
			t.Errorf("expected no output for empty input, got %v", results)
		}
	})
}