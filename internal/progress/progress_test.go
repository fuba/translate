package progress

import (
	"bytes"
	"strings"
	"testing"
)

func TestReporterTickAndDone(t *testing.T) {
	var buf bytes.Buffer
	r := New(&buf, WithMinInterval(0))
	r.Tick("")
	r.Tick("")
	r.Done()

	out := buf.String()
	if !strings.Contains(out, "translating... 1") {
		t.Fatalf("expected first update, got %q", out)
	}
	if !strings.Contains(out, "translating... 2") {
		t.Fatalf("expected second update, got %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("expected trailing newline, got %q", out)
	}
}

func TestReporterWithTotal(t *testing.T) {
	var buf bytes.Buffer
	r := New(&buf, WithMinInterval(0))
	r.SetTotal(3)
	r.Tick("")
	r.Tick("")
	r.Done()

	out := buf.String()
	if !strings.Contains(out, "translating... 1/3 (33%)") {
		t.Fatalf("expected percent update, got %q", out)
	}
	if !strings.Contains(out, "translating... 2/3 (66%)") {
		t.Fatalf("expected percent update, got %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("expected trailing newline, got %q", out)
	}
}
