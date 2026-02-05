package pdf

import (
	"strings"

	"github.com/unidoc/unipdf/v4/extractor"
	"github.com/unidoc/unipdf/v4/model"
)

type textLine struct {
	Text string
	Box  model.PdfRectangle
}

func groupLines(marks []extractor.TextMark) []textLine {
	lines := make([]textLine, 0)
	var current textLine
	var hasBox bool

	flush := func() {
		if strings.TrimSpace(current.Text) == "" {
			current = textLine{}
			hasBox = false
			return
		}
		lines = append(lines, current)
		current = textLine{}
		hasBox = false
	}

	for _, m := range marks {
		if m.Meta {
			if strings.Contains(m.Text, "\n") {
				flush()
				continue
			}
			current.Text += m.Text
			continue
		}

		current.Text += m.Text
		if !hasBox {
			current.Box = m.BBox
			hasBox = true
		} else {
			current.Box = unionBox(current.Box, m.BBox)
		}
	}
	flush()
	return lines
}

func unionBox(a, b model.PdfRectangle) model.PdfRectangle {
	return model.PdfRectangle{
		Llx: minFloat(a.Llx, b.Llx),
		Lly: minFloat(a.Lly, b.Lly),
		Urx: maxFloat(a.Urx, b.Urx),
		Ury: maxFloat(a.Ury, b.Ury),
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
