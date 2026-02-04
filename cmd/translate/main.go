package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fuba/translate/internal/app"
)

func main() {
	var cfg app.Config

	flag.StringVar(&cfg.Format, "format", "auto", "input format: text|md|pdf|auto")
	flag.StringVar(&cfg.InPath, "in", "", "input path (default: stdin)")
	flag.StringVar(&cfg.OutPath, "out", "", "output path (default: stdout)")
	flag.StringVar(&cfg.From, "from", "auto", "source language code (default: auto)")
	flag.StringVar(&cfg.To, "to", "", "target language code (default: from LANG)")
	flag.StringVar(&cfg.Model, "model", "gpt-oss-20b", "model name")
	flag.StringVar(&cfg.BaseURL, "base-url", "http://kirgizu:8080", "OpenAI compatible base URL")
	flag.StringVar(&cfg.APIKey, "api-key", os.Getenv("OPENAI_API_KEY"), "API key (default: OPENAI_API_KEY)")
	flag.DurationVar(&cfg.Timeout, "timeout", 120*time.Second, "HTTP timeout")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "translate - translate text/markdown/pdf via OpenAI compatible API\n\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintln(os.Stderr, "  translate --from en --to ja --in input.txt --out output.txt")
		fmt.Fprintln(os.Stderr, "  cat input.md | translate --format md --to ja > output.md")
		fmt.Fprintln(os.Stderr, "  translate --format pdf --in input.pdf --out output.pdf")
	}

	flag.Parse()

	if err := app.Run(context.Background(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
