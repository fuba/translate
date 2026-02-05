package pdf

import "testing"

func TestJoinPageText(t *testing.T) {
	pages := []string{"a", "b", ""}
	got := joinPageText(pages)
	want := "=== Page 1 ===\na\n\n=== Page 2 ===\nb\n\n=== Page 3 ===\n\n\n"
	if got != want {
		t.Fatalf("joinPageText got %q want %q", got, want)
	}
}
