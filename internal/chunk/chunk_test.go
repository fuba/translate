package chunk

import "testing"

func TestSplitPrefersPunctuation(t *testing.T) {
	text := "Hello world. This is a test. Next"
	chunks := Split(text, 12)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %v", chunks)
	}
	if chunks[0] != "Hello world." {
		t.Fatalf("first chunk = %q", chunks[0])
	}
}

func TestSplitKeepsAllText(t *testing.T) {
	text := "あいうえおかきくけこさしすせそ"
	chunks := Split(text, 5)
	joined := ""
	for _, c := range chunks {
		joined += c
	}
	if joined != text {
		t.Fatalf("joined = %q, want %q", joined, text)
	}
}

func TestSplitNoMax(t *testing.T) {
	text := "short"
	chunks := Split(text, 0)
	if len(chunks) != 1 || chunks[0] != text {
		t.Fatalf("unexpected chunks: %v", chunks)
	}
}
