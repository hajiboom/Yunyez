// Package buffer provides the audio buffer to control the publish rate
// a text buffer as the pre handler for LLM response before tts
package buffer

import (
	"strings"
	"time"
	"unicode"
)


var (
	punctuation = "。！？；;,.!?，：:、\n\t"
	textMaxLength = 45 // the max length of text
	textMinLength = 12 // the min length of text
	// the delay time to flush the text buffer while no punctuation received
	flushDelay = 600 * time.Millisecond 
)

// TextBuffer is the text buffer to preprocess the text from llm
type TextBuffer struct {
	input <-chan string // the input channel (only read) to receive the text
	output chan string
}

func NewTextBuffer(input <-chan string) *TextBuffer {
	tb := &TextBuffer{
		input:  input,
		output: make(chan string, 4),
	}
	go tb.run()
	return tb
}


// Output returns the output channel (only read) to send the text to tts
func (tb *TextBuffer) Output() <-chan string {
	return tb.output
}

func (tb *TextBuffer) run() {
	defer close(tb.output)

	var buffer strings.Builder
	timer := time.NewTimer(flushDelay)
	defer timer.Stop()

	for {
		select {
		case part, ok := <-tb.input:
			if !ok {
				// llm streaming end
				if buffer.Len() > 0 {
					tb.output <- buffer.String()
				}
				return
			}

			buffer.WriteString(part)
			text := buffer.String()

			// rule 1: try to find the safe-split position, if the length is greater than min length
			if len(text) >= textMinLength {
				pos := -1
				runes := []rune(text)
				start := len(runes) - 1
				end := len(runes)
				if end - start > 10 {
					start = end - 10
				}

				for i := end -1; i >= start; i-- {
					if safeBreakPosition(runes[i]) {
						pos = i + 1 // the next position after the break position
						break
					}
				}

				if pos > 0 && pos < len(runes) {
					// find out the safe break position but not all text
					sentence := strings.TrimSpace(string(runes[:pos]))
					if sentence != "" {
						tb.output <- sentence
						buffer.Reset() // save the rest partition
						buffer.WriteString(string(runes[pos:]))
						continue
					}
				}
			}
			// rule 2: strictly splitting the partition beyond the max length
			//  although it's not a safe break position
			if buffer.Len() > textMaxLength {
				tb.output <- buffer.String()
				buffer.Reset()
			}
			// reset timer
			if !timer.Stop() {
				select {
				case <-timer.C: // drain if not empty
				default:
				}
			}
			timer.Reset(flushDelay)
		
		case <-timer.C: // timeout: flush current text although it's patchy
			if buffer.Len() >= 8 {
				tb.output <- buffer.String()
				buffer.Reset()
			}
			timer.Reset(flushDelay)
		}
	}
}

// safeBreakPosition checks if the position is a safe break position
// params:
// - r: the rune to check
// return: 
// - bool: true if the position is a safe break position
func safeBreakPosition(r rune) bool {
	// chinese character / punctuation / space
	if unicode.Is(unicode.Han, r) {
		return true
	}
	if strings.ContainsRune(punctuation, r) {
		return true
	}
	if unicode.IsSpace(r) {
		return true
	}

	return false

}
