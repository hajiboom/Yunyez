// Package buffer provides the audio buffer to control the publish rate
// a text buffer as the pre handler for LLM response before tts
package buffer

import (
	"regexp"
	"strings"
	"time"
)


var (
	sentenceEnd = regexp.MustCompile(`[。！？.!?;；…]`)
	textMaxLength = 50 // the max length of text
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
			// rule 1: split the text by sentence end
			if sentenceEnd.MatchString(part) {
				lastEnd := sentenceEnd.FindAllStringIndex(text, -1)
				if len(lastEnd) > 0 {
					endPos := lastEnd[len(lastEnd)-1][1]
					sentence := strings.TrimSpace(text[:endPos])
					if sentence != "" {
						tb.output <- sentence
					}
					// 保留剩余部分
					buffer.Reset()
					remaining := text[endPos:]
					if remaining != "" {
						buffer.WriteString(remaining)
					}
				}
			}
			// rule 2: beyond the max length
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
			if buffer.Len() > 0 {
				tb.output <- buffer.String()
				buffer.Reset()
			}
			timer.Reset(flushDelay)
		}
	}
}
