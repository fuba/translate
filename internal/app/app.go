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

	"github.com/fuba/translate/internal/lang"
	"github.com/fuba/translate/internal/llm"
	"github.com/fuba/translate/internal/markdown"
	"github.com/fuba/translate/internal/pdf"
)

type Config struct {
	Format  string
	InPath  string
	OutPath string
	From    string
	To      string
	Model   string
	BaseURL string
	APIKey  string
	Timeout time.Duration
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

	client, err := llm.NewClient(cfg.BaseURL, cfg.Model, llm.WithAPIKey(cfg.APIKey), llm.WithTimeout(cfg.Timeout))
	if err != nil {
		return err
	}

	switch format {
	case "text":
		input, err := readInput(cfg.InPath)
		if err != nil {
			return err
		}
		out, err := client.Translate(ctx, string(input), cfg.From, cfg.To, "text")
		if err != nil {
			return err
		}
		return writeOutput(cfg.OutPath, []byte(out))
	case "md":
		input, err := readInput(cfg.InPath)
		if err != nil {
			return err
		}
		out, err := markdown.Translate(ctx, client, input, cfg.From, cfg.To)
		if err != nil {
			return err
		}
		return writeOutput(cfg.OutPath, out)
	case "pdf":
		if cfg.InPath == "" || cfg.InPath == "-" {
			return errors.New("pdf input requires a file path")
		}
		if cfg.OutPath == "" || cfg.OutPath == "-" {
			return errors.New("pdf output requires a file path")
		}
		return pdf.Translate(ctx, client, cfg.InPath, cfg.OutPath, cfg.From, cfg.To)
	default:
		return fmt.Errorf("unsupported format: %s", format)
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
