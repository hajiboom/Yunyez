// 测试音频协议的序列化和反序列化
package voice

import (
	"testing"
)

func TestHeaderRoundTrip(t *testing.T) {
	original := &Header{
		Version:    1,
		AudioFormat:  3,        // OPUS
		SampleRate: 16000,
		Ch:         1,        // mono
		F:          1,        // full frame
		FrameSeq:   42,
		Timestamp:  1000,
		PayloadLen: 28,
		CRC16:      0x1234,
	}

	// Marshal
	data := original.Marshal()

	if len(data) != HeaderSize {
		t.Fatalf("Expected header size %d, got %d", HeaderSize, len(data))
	}

	// Unmarshal
	recovered, err := UnmarshalHeader(data)
	if err != nil {
		t.Fatal("Unmarshal failed:", err)
	}

	// Compare
	if *original != *recovered {
		t.Errorf("Round-trip mismatch!\nOriginal: %+v\nRecovered: %+v", original, recovered)
	}

	// Also test bit fields explicitly
	if recovered.Version != 1 ||
		recovered.AudioFormat != 3 ||
		recovered.SampleRate != 16000 ||
		recovered.Ch != 1 ||
		recovered.F != 1 {
		t.Error("Field values incorrect after round-trip")
	}
}

// Helper: dummy crc_16 for testing
func crc_16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func TestFullPacketWithCRC(t *testing.T) {
	payload := []byte("OPUS_FRAME_1234567890")
	header := &Header{
		Version:    1,
		AudioFormat:  3,
		SampleRate: 16000,
		Ch:         1,
		F:          1,
		FrameSeq:   100,
		Timestamp:  2000,
		PayloadLen: uint16(len(payload)),
		CRC16:      0, // will compute
	}

	// Step 1: marshal header (CRC=0)
	rawHeader := header.Marshal()
	// Ensure CRC field is zero for checksum
	rawHeader[10] = 0
	rawHeader[11] = 0

	// Step 2: compute CRC over [header(without CRC)][payload]
	checksumData := append(rawHeader, payload...)
	expectedCRC := crc_16(checksumData)

	// Step 3: set CRC and build final packet
	header.CRC16 = expectedCRC
	finalHeader := header.Marshal()
	packet := append(finalHeader, payload...)

	// --- Receiver side simulation ---
	// Extract header from packet
	receivedHeader, err := UnmarshalHeader(packet[:HeaderSize])
	if err != nil {
		t.Fatal(err)
	}

	// Reconstruct the data that was used to compute CRC:
	// i.e., original header with CRC=0 + payload
	dataForCRC := make([]byte, len(packet))
	copy(dataForCRC, packet)           // copy full packet
	dataForCRC[10] = 0                 // zero out CRC high byte
	dataForCRC[11] = 0                 // zero out CRC low byte

	computedCRC := crc_16(dataForCRC)

	if receivedHeader.CRC16 != computedCRC {
		t.Errorf("CRC mismatch: got %04x, expected %04x", receivedHeader.CRC16, computedCRC)
	}
}