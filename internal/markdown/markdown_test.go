package markdown

import (
	"context"
	"strings"
	"testing"
)

type upperTranslator struct{}

func (upperTranslator) Translate(ctx context.Context, text, from, to, format string) (string, error) {
	_ = ctx
	_ = from
	_ = to
	_ = format
	return strings.ToUpper(text), nil
}

func TestTranslateMarkdown(t *testing.T) {
	input := "# Title\n\nParagraph with **bold** and `code`.\n\n- item1\n- item2\n\n```go\nfmt.Println(\"hello\")\n```\n\n[link text](https://example.com)\n"
	got, err := Translate(context.Background(), upperTranslator{}, []byte(input), "en", "ja")
	if err != nil {
		t.Fatalf("Translate error: %v", err)
	}

	out := string(got)
	if !strings.Contains(out, "# TITLE") {
		t.Fatalf("expected title to be translated: %q", out)
	}
	if !strings.Contains(out, "PARAGRAPH WITH ") {
		t.Fatalf("expected paragraph to be translated: %q", out)
	}
	if !strings.Contains(out, "**BOLD**") {
		t.Fatalf("expected bold text to be translated: %q", out)
	}
	if !strings.Contains(out, "`code`") {
		t.Fatalf("expected inline code to remain unchanged: %q", out)
	}
	if !strings.Contains(out, "fmt.Println(\"hello\")") {
		t.Fatalf("expected code block to remain unchanged: %q", out)
	}
	if !strings.Contains(out, "[LINK TEXT](https://example.com)") {
		t.Fatalf("expected link text to be translated and URL preserved: %q", out)
	}
}

func TestTranslateMarkdownProgress(t *testing.T) {
	input := "Hello **world**.\n\nSecond line."
	var got []string
	_, err := TranslateWithProgress(context.Background(), upperTranslator{}, []byte(input), "en", "ja",
		func(text string) {
			got = append(got, text)
		})
	if err != nil {
		t.Fatalf("TranslateWithProgress error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected progress callbacks, got none")
	}
}
