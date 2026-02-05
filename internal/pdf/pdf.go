package pdf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fuba/translate/internal/chunk"
	"github.com/fuba/translate/internal/translate"
	"github.com/unidoc/unipdf/v4/common/license"
	"github.com/unidoc/unipdf/v4/creator"
	"github.com/unidoc/unipdf/v4/extractor"
	"github.com/unidoc/unipdf/v4/model"
)

func Translate(ctx context.Context, tr translate.Translator, inPath, outPath, from, to, unidocKey string, maxChars int, progress func(string), fontPath string) error {
	if strings.TrimSpace(unidocKey) == "" {
		return errors.New("unidoc key is required for PDF translation")
	}
	if err := setLicense(unidocKey); err != nil {
		return fmt.Errorf("set unidoc license: %w", err)
	}

	f, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return err
	}

	encrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return err
	}
	if encrypted {
		ok, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("encrypted pdf requires password")
		}
	}

	count, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	c := creator.New()
	overlayFont, err := loadOverlayFont(fontPath)
	if err != nil {
		return err
	}

	for pageNum := 1; pageNum <= count; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return err
		}
		if err := c.AddPage(page); err != nil {
			return err
		}
		if progress != nil {
			progress(fmt.Sprintf("[page %d] translating", pageNum))
		}
		if err := overlayTranslatedLines(ctx, tr, c, page, from, to, maxChars, progress, overlayFont); err != nil {
			return fmt.Errorf("page %d: %w", pageNum, err)
		}
	}

	return c.WriteToFile(outPath)
}

func ExtractText(inPath, unidocKey string) (string, error) {
	if strings.TrimSpace(unidocKey) == "" {
		return "", errors.New("unidoc key is required for PDF extraction")
	}
	if err := setLicense(unidocKey); err != nil {
		return "", fmt.Errorf("set unidoc license: %w", err)
	}

	f, err := os.Open(inPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return "", err
	}

	encrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return "", err
	}
	if encrypted {
		ok, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return "", err
		}
		if !ok {
			return "", errors.New("encrypted pdf requires password")
		}
	}

	count, err := pdfReader.GetNumPages()
	if err != nil {
		return "", err
	}

	pages := make([]string, 0, count)
	for pageNum := 1; pageNum <= count; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return "", err
		}
		ex, err := extractor.New(page)
		if err != nil {
			return "", err
		}
		pageText, _, _, err := ex.ExtractPageText()
		if err != nil {
			return "", err
		}
		pages = append(pages, pageText.Text())
	}

	return joinPageText(pages), nil
}

func CountChunks(inPath, unidocKey string, maxChars int) (int, error) {
	if strings.TrimSpace(unidocKey) == "" {
		return 0, errors.New("unidoc key is required for PDF extraction")
	}
	if err := setLicense(unidocKey); err != nil {
		return 0, fmt.Errorf("set unidoc license: %w", err)
	}

	f, err := os.Open(inPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return 0, err
	}

	encrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return 0, err
	}
	if encrypted {
		ok, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, errors.New("encrypted pdf requires password")
		}
	}

	count, err := pdfReader.GetNumPages()
	if err != nil {
		return 0, err
	}

	total := 0
	for pageNum := 1; pageNum <= count; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return 0, err
		}
		extract, err := extractor.New(page)
		if err != nil {
			return 0, err
		}
		pageText, _, _, err := extract.ExtractPageText()
		if err != nil {
			return 0, err
		}
		lines := groupLines(pageText.Marks().Elements())
		for _, line := range lines {
			if strings.TrimSpace(line.Text) == "" {
				continue
			}
			total += len(chunk.Split(line.Text, maxChars))
		}
	}

	return total, nil
}

func setLicense(key string) error {
	if err := license.SetMeteredKey(key); err != nil {
		if isLicenseAlreadySet(err) {
			return nil
		}
		return err
	}
	return nil
}

func isLicenseAlreadySet(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "license key already set")
}

func overlayTranslatedLines(ctx context.Context, tr translate.Translator, c *creator.Creator, page *model.PdfPage, from, to string, maxChars int, progress func(string), font *model.PdfFont) error {
	ex, err := extractor.New(page)
	if err != nil {
		return err
	}
	pageText, _, _, err := ex.ExtractPageText()
	if err != nil {
		return err
	}
	mediaBox, err := page.GetMediaBox()
	if err != nil {
		return err
	}
	if page.MediaBox == nil {
		page.MediaBox = mediaBox
	}
	pageHeight := mediaBox.Ury

	lines := groupLines(pageText.Marks().Elements())
	for _, line := range lines {
		if strings.TrimSpace(line.Text) == "" {
			continue
		}
		translated, err := translateChunked(ctx, tr, line.Text, from, to, maxChars, progress)
		if err != nil {
			return err
		}
		drawLineOverlay(c, line, translated, pageHeight, font)
	}
	return nil
}

func drawLineOverlay(c *creator.Creator, line textLine, translated string, pageHeight float64, font *model.PdfFont) {
	y := pageHeight - line.Box.Ury
	height := line.Box.Ury - line.Box.Lly
	width := line.Box.Urx - line.Box.Llx

	rect := c.NewRectangle(line.Box.Llx, y, width, height)
	rect.SetFillColor(creator.ColorWhite)
	rect.SetBorderColor(creator.ColorWhite)
	_ = c.Draw(rect)

	p := c.NewStyledParagraph()
	p.SetText(translated)
	p.SetPos(line.Box.Llx, y)
	p.SetFontSize(height)
	if font != nil {
		p.SetFont(font)
	}
	_ = c.Draw(p)
}

func loadOverlayFont(fontPath string) (*model.PdfFont, error) {
	if strings.TrimSpace(fontPath) == "" {
		return model.NewStandard14Font("Helvetica")
	}
	if _, err := os.Stat(fontPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("pdf font not found: %s", fontPath)
		}
		return nil, err
	}
	return model.NewCompositePdfFontFromTTFFile(fontPath)
}

func translateChunked(ctx context.Context, tr translate.Translator, text, from, to string, maxChars int, progress func(string)) (string, error) {
	parts := chunk.Split(text, maxChars)
	var b strings.Builder
	for _, part := range parts {
		out, err := tr.Translate(ctx, part, from, to, "text")
		if err != nil {
			return "", err
		}
		b.WriteString(out)
		if progress != nil {
			progress(out)
		}
	}
	return b.String(), nil
}

func joinPageText(pages []string) string {
	var b strings.Builder
	for i, text := range pages {
		b.WriteString(fmt.Sprintf("=== Page %d ===\n", i+1))
		b.WriteString(text)
		b.WriteString("\n\n")
	}
	return b.String()
}
