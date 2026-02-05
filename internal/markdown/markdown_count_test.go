package markdown

import "testing"

func TestCountChunks(t *testing.T) {
	input := []byte("Hello **world**\n\n```\ncode block\n```\n\nMore text.")
	total := CountChunks(input, 5)
	if total == 0 {
		t.Fatalf("expected chunks, got %d", total)
	}
}
