package pdf

import (
	"testing"

	"github.com/unidoc/unipdf/v4/extractor"
	"github.com/unidoc/unipdf/v4/model"
)

func TestGroupLines(t *testing.T) {
	marks := []extractor.TextMark{
		{Text: "Hello", BBox: model.PdfRectangle{Llx: 10, Lly: 10, Urx: 30, Ury: 20}},
		{Text: " ", Meta: true},
		{Text: "world", BBox: model.PdfRectangle{Llx: 35, Lly: 10, Urx: 60, Ury: 20}},
		{Text: "\n", Meta: true},
		{Text: "Next", BBox: model.PdfRectangle{Llx: 10, Lly: 30, Urx: 30, Ury: 40}},
	}

	lines := groupLines(marks)
	if len(lines) != 2 {
		t.Fatalf("lines=%d", len(lines))
	}
	if lines[0].Text != "Hello world" {
		t.Fatalf("line0 text=%q", lines[0].Text)
	}
	if lines[1].Text != "Next" {
		t.Fatalf("line1 text=%q", lines[1].Text)
	}
}
