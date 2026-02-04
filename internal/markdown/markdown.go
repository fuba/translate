package markdown

import (
	"context"
	"strings"

	"github.com/fuba/translate/internal/translate"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

type textSegment struct {
	start int
	stop  int
	text  string
}

type ProgressFunc func(text string)

func Translate(ctx context.Context, tr translate.Translator, input []byte, from, to string) ([]byte, error) {
	return TranslateWithProgress(ctx, tr, input, from, to, nil)
}

func TranslateWithProgress(ctx context.Context, tr translate.Translator, input []byte, from, to string, progress ProgressFunc) ([]byte, error) {
	segments := collectTextSegments(input)
	if len(segments) == 0 {
		return append([]byte(nil), input...), nil
	}

	translated := make([]string, len(segments))
	for i, seg := range segments {
		if strings.TrimSpace(seg.text) == "" {
			translated[i] = seg.text
			continue
		}
		out, err := tr.Translate(ctx, seg.text, from, to, "text")
		if err != nil {
			return nil, err
		}
		translated[i] = out
		if progress != nil {
			progress(out)
		}
	}

	out := append([]byte(nil), input...)
	for i := len(segments) - 1; i >= 0; i-- {
		seg := segments[i]
		repl := []byte(translated[i])
		out = append(out[:seg.start], append(repl, out[seg.stop:]...)...)
	}
	return out, nil
}

func collectTextSegments(input []byte) []textSegment {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Table,
			extension.TaskList,
			extension.Linkify,
		),
	)

	reader := text.NewReader(input)
	doc := md.Parser().Parse(reader)

	segments := make([]textSegment, 0, 64)
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		textNode, ok := n.(*ast.Text)
		if !ok {
			return ast.WalkContinue, nil
		}
		if textNode.IsRaw() {
			return ast.WalkContinue, nil
		}
		if !shouldTranslateNode(n) {
			return ast.WalkContinue, nil
		}

		seg := textNode.Segment
		if seg.Start >= seg.Stop {
			return ast.WalkContinue, nil
		}
		segments = append(segments, textSegment{
			start: seg.Start,
			stop:  seg.Stop,
			text:  string(seg.Value(input)),
		})
		return ast.WalkContinue, nil
	})

	return segments
}

func shouldTranslateNode(n ast.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p.(type) {
		case *ast.CodeBlock,
			*ast.FencedCodeBlock,
			*ast.CodeSpan,
			*ast.HTMLBlock,
			*ast.RawHTML,
			*ast.AutoLink:
			return false
		}
	}
	return true
}
