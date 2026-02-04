package pdf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fuba/translate/internal/translate"
	"github.com/unidoc/unipdf/v4/common"
	"github.com/unidoc/unipdf/v4/common/license"
	"github.com/unidoc/unipdf/v4/contentstream"
	"github.com/unidoc/unipdf/v4/core"
	"github.com/unidoc/unipdf/v4/model"
	"github.com/unidoc/unipdf/v4/model/optimize"
)

func Translate(ctx context.Context, tr translate.Translator, inPath, outPath, from, to, unidocKey string) error {
	if strings.TrimSpace(unidocKey) == "" {
		return errors.New("unidoc key is required for PDF translation")
	}
	if err := license.SetMeteredKey(unidocKey); err != nil {
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

	pdfWriter := model.NewPdfWriter()
	count, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	for pageNum := 1; pageNum <= count; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return err
		}
		if err := translatePageText(ctx, tr, page, from, to); err != nil {
			return fmt.Errorf("page %d: %w", pageNum, err)
		}
		if err := pdfWriter.AddPage(page); err != nil {
			return err
		}
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	opt := optimize.Options{
		CombineDuplicateStreams:         true,
		CombineIdenticalIndirectObjects: true,
		UseObjectStreams:                true,
		CompressStreams:                 true,
	}
	pdfWriter.SetOptimizer(optimize.New(opt))

	return pdfWriter.Write(outFile)
}

func translatePageText(ctx context.Context, tr translate.Translator, page *model.PdfPage, from, to string) error {
	contents, err := page.GetAllContentStreams()
	if err != nil {
		return err
	}

	ops, err := contentstream.NewContentStreamParser(contents).Parse()
	if err != nil {
		return err
	}

	var currFont *model.PdfFont
	cache := map[string]string{}

	translateObj := func(objptr *core.PdfObject) error {
		strObj, ok := core.GetString(*objptr)
		if !ok {
			common.Log.Debug("Invalid parameter, skipping")
			return nil
		}

		decoded := strObj.String()
		if currFont != nil {
			val, _, numMisses := currFont.CharcodeBytesToUnicode(strObj.Bytes())
			if numMisses != 0 {
				common.Log.Debug("WARN: some charcodes could not be decoded")
			}
			decoded = val
		}

		if strings.TrimSpace(decoded) == "" {
			return nil
		}

		translated, ok := cache[decoded]
		if !ok {
			out, err := tr.Translate(ctx, decoded, from, to, "text")
			if err != nil {
				return err
			}
			translated = out
			cache[decoded] = translated
		}

		if currFont != nil {
			encodedBytes, numMisses := currFont.StringToCharcodeBytes(translated)
			if numMisses != 0 {
				common.Log.Debug("WARN: some runes could not be encoded")
			}
			*strObj = *core.MakeString(string(encodedBytes))
			return nil
		}

		*strObj = *core.MakeString(translated)
		return nil
	}

	processor := contentstream.NewContentStreamProcessor(*ops)
	processor.AddHandler(contentstream.HandlerConditionEnumAllOperands, "",
		func(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
			switch op.Operand {
			case "Tj", "'":
				if len(op.Params) != 1 {
					common.Log.Debug("Invalid: Tj/' with invalid params")
					return nil
				}
				return translateObj(&op.Params[0])
			case "\"":
				if len(op.Params) < 1 {
					common.Log.Debug("Invalid: \" with invalid params")
					return nil
				}
				idx := len(op.Params) - 1
				return translateObj(&op.Params[idx])
			case "TJ":
				if len(op.Params) != 1 {
					common.Log.Debug("Invalid: TJ with invalid params")
					return nil
				}
				arr, _ := core.GetArray(op.Params[0])
				for i := range arr.Elements() {
					obj := arr.Get(i)
					if err := translateObj(&obj); err != nil {
						return err
					}
					arr.Set(i, obj)
				}
				return nil
			case "Tf":
				if len(op.Params) != 2 {
					common.Log.Debug("Invalid: Tf with invalid params")
					return nil
				}
				fname, ok := core.GetName(op.Params[0])
				if !ok || fname == nil {
					common.Log.Debug("ERROR: could not get font name")
					return nil
				}

				fObj, has := resources.GetFontByName(*fname)
				if !has {
					common.Log.Debug("ERROR: font %s not found", fname.String())
					return nil
				}

				pdfFont, err := model.NewPdfFontFromPdfObject(fObj)
				if err != nil {
					common.Log.Debug("ERROR: loading font")
					return nil
				}
				currFont = pdfFont
				return nil
			default:
				return nil
			}
		})

	if err := processor.Process(page.Resources); err != nil {
		return err
	}

	return page.SetContentStreams([]string{ops.String()}, core.NewFlateEncoder())
}
