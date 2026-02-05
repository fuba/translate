package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fuba/translate/internal/chunk"
	"github.com/fuba/translate/internal/lang"
	"github.com/fuba/translate/internal/llm"
	"github.com/fuba/translate/internal/markdown"
	progressui "github.com/fuba/translate/internal/progress"
	"github.com/fuba/translate/internal/pdf"
	"github.com/fuba/translate/internal/secure"
	"github.com/fuba/translate/internal/translate"
	"golang.org/x/term"
)

type Config struct {
	Format        string
	InPath        string
	OutPath       string
	From          string
	To            string
	Model         string
	BaseURL       string
	APIKey        string
	Timeout       time.Duration
	Verbose       bool
	Silent        bool
	MaxChars      int
	Endpoint      string
	PassphraseTTL time.Duration
	DumpExtracted string
	VerbosePrompt bool
	PDFFont       string
}

func Run(ctx context.Context, cfg Config) error {
	format, err := resolveFormat(cfg.Format, cfg.InPath)
	if err != nil {
		return err
	}

	if strings.TrimSpace(cfg.To) == "" {
		cfg.To = lang.DefaultTargetLang(os.Getenv("LANG"))
	}
	if strings.TrimSpace(cfg.To) == "" {
		return errors.New("target language is required")
	}

	client, err := llm.NewClient(
		cfg.BaseURL,
		cfg.Model,
		llm.WithAPIKey(cfg.APIKey),
		llm.WithTimeout(cfg.Timeout),
		llm.WithEndpoint(cfg.Endpoint),
		llm.WithDebugLogger(promptLogger(cfg.VerbosePrompt)),
	)
	if err != nil {
		return err
	}

	writesToStdout := cfg.OutPath == "" || cfg.OutPath == "-"
	if strings.TrimSpace(cfg.DumpExtracted) != "" {
		writesToStdout = cfg.DumpExtracted == "-"
	}

	progressFn := func(string) {}
	var reporter *progressui.Reporter
	if cfg.Silent {
		progressFn = func(string) {}
		reporter = nil
	} else if cfg.Verbose {
		progressFn = func(text string) {
			fmt.Fprintln(os.Stderr, text)
		}
	} else if !writesToStdout && term.IsTerminal(int(os.Stderr.Fd())) {
		reporter = progressui.New(os.Stderr)
		progressFn = func(text string) {
			if strings.HasPrefix(text, "[page ") && strings.HasSuffix(text, "] translating") {
				return
			}
			reporter.Tick(text)
		}
	}
	if reporter != nil {
		defer reporter.Done()
	}

	switch format {
	case "text":
		input, err := readInput(cfg.InPath)
		if err != nil {
			return err
		}
		if reporter != nil {
			reporter.SetTotal(len(chunk.Split(string(input), cfg.MaxChars)))
		}
		out, err := translateText(ctx, client, string(input), cfg.From, cfg.To, cfg.MaxChars, progressFn)
		if err != nil {
			return err
		}
		return writeOutput(cfg.OutPath, []byte(out))
	case "md":
		input, err := readInput(cfg.InPath)
		if err != nil {
			return err
		}
		if reporter != nil {
			reporter.SetTotal(markdown.CountChunks(input, cfg.MaxChars))
		}
		out, err := markdown.TranslateWithProgress(ctx, client, input, cfg.From, cfg.To, cfg.MaxChars, progressFn)
		if err != nil {
			return err
		}
		return writeOutput(cfg.OutPath, out)
	case "pdf":
		if cfg.InPath == "" || cfg.InPath == "-" {
			return errors.New("pdf input requires a file path")
		}
		if strings.TrimSpace(cfg.DumpExtracted) != "" {
			unidocKey, err := secure.LoadUnidocKey(cfg.PassphraseTTL)
			if err != nil {
				return err
			}
			text, err := pdf.ExtractText(cfg.InPath, unidocKey)
			if err != nil {
				return err
			}
			return writeOutput(cfg.DumpExtracted, []byte(text))
		}
		if cfg.OutPath == "" || cfg.OutPath == "-" {
			return errors.New("pdf output requires a file path")
		}
		unidocKey, err := secure.LoadUnidocKey(cfg.PassphraseTTL)
		if err != nil {
			return err
		}
		if reporter != nil {
			total, err := pdf.CountChunks(cfg.InPath, unidocKey, cfg.MaxChars)
			if err != nil {
				return err
			}
			reporter.SetTotal(total)
		}
		return pdf.Translate(ctx, client, cfg.InPath, cfg.OutPath, cfg.From, cfg.To, unidocKey, cfg.MaxChars, progressFn, cfg.PDFFont)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func translateText(ctx context.Context, tr translate.Translator, text, from, to string, maxChars int, progress func(string)) (string, error) {
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

func promptLogger(enabled bool) func(string) {
	if !enabled {
		return nil
	}
	return func(msg string) {
		fmt.Fprintln(os.Stderr, msg)
	}
}

func resolveFormat(format, inPath string) (string, error) {
	f := strings.ToLower(strings.TrimSpace(format))
	if f == "" || f == "auto" {
		return detectFormatFromPath(inPath), nil
	}

	switch f {
	case "text", "md", "markdown", "pdf":
		if f == "markdown" {
			return "md", nil
		}
		return f, nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

func detectFormatFromPath(inPath string) string {
	if inPath == "" || inPath == "-" {
		return "text"
	}
	ext := strings.ToLower(filepath.Ext(inPath))
	switch ext {
	case ".md", ".markdown":
		return "md"
	case ".pdf":
		return "pdf"
	default:
		return "text"
	}
}

func readInput(path string) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func writeOutput(path string, data []byte) error {
	if path == "" || path == "-" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
