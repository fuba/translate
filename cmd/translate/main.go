package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fuba/translate/internal/app"
	"github.com/fuba/translate/internal/config"
	"github.com/fuba/translate/internal/secure"
	"golang.org/x/term"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "auth":
			if err := runAuth(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		case "config":
			if err := runConfig(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	cfgFile, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		os.Exit(1)
	}

	var cfg app.Config

	flag.StringVar(&cfg.Format, "format", config.StringOrFallback(cfgFile.Format, "auto"), "input format: text|md|pdf|auto")
	flag.StringVar(&cfg.InPath, "in", "", "input path (default: stdin)")
	flag.StringVar(&cfg.OutPath, "out", "", "output path (default: stdout)")
	flag.StringVar(&cfg.From, "from", config.StringOrFallback(cfgFile.From, "auto"), "source language code (default: auto)")
	flag.StringVar(&cfg.To, "to", cfgFile.To, "target language code (default: from LANG)")
	flag.StringVar(&cfg.Model, "model", config.StringOrFallback(cfgFile.Model, "gpt-oss-20b"), "model name")
	flag.StringVar(&cfg.BaseURL, "base-url", config.StringOrFallback(cfgFile.BaseURL, "http://kirgizu:8080"), "OpenAI compatible base URL")
	flag.StringVar(&cfg.APIKey, "api-key", os.Getenv("OPENAI_API_KEY"), "API key (default: OPENAI_API_KEY)")
	flag.DurationVar(&cfg.Timeout, "timeout", config.Timeout(cfgFile, 120*time.Second), "HTTP timeout")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "print translated chunks to stderr")
	flag.IntVar(&cfg.MaxChars, "max-chars", config.IntOrFallback(cfgFile.MaxChars, 2000), "max chars per translation request (0 disables)")
	flag.StringVar(&cfg.Endpoint, "endpoint", config.StringOrFallback(cfgFile.Endpoint, "completion"), "endpoint: chat|completion|auto")
	flag.DurationVar(&cfg.PassphraseTTL, "passphrase-ttl", config.PassphraseTTL(cfgFile, 10*time.Minute), "cache passphrase for duration (0 disables)")
	flag.StringVar(&cfg.DumpExtracted, "dump-extracted", "", "dump raw extracted PDF text to path (use - for stdout)")
	flag.BoolVar(&cfg.VerbosePrompt, "verbose-prompt", false, "print prompts to stderr")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "translate - translate text/markdown/pdf via OpenAI compatible API\n\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintln(os.Stderr, "  translate --from en --to ja --in input.txt --out output.txt")
		fmt.Fprintln(os.Stderr, "  cat input.md | translate --format md --to ja > output.md")
		fmt.Fprintln(os.Stderr, "  translate --format pdf --in input.pdf --out output.pdf")
		fmt.Fprintln(os.Stderr, "\nConfig:")
		fmt.Fprintln(os.Stderr, "  translate config set --base-url http://kirgizu:8080 --model gpt-oss-20b")
		fmt.Fprintln(os.Stderr, "\nSecrets:")
		fmt.Fprintln(os.Stderr, "  translate auth set-unidoc")
	}

	flag.Parse()

	if err := app.Run(context.Background(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runAuth(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("auth subcommand is required")
	}
	switch args[0] {
	case "set-unidoc":
		return authSetUnidoc(args[1:])
	default:
		return fmt.Errorf("unknown auth subcommand: %s", args[0])
	}
}

func authSetUnidoc(args []string) error {
	fs := flag.NewFlagSet("auth set-unidoc", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	if err := fs.Parse(args); err != nil {
		return err
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("auth set-unidoc requires a TTY")
	}

	key, err := promptHidden("UNIDOC license key: ")
	if err != nil {
		return err
	}
	pass1, err := promptHidden("Passphrase: ")
	if err != nil {
		return err
	}
	pass2, err := promptHidden("Passphrase (again): ")
	if err != nil {
		return err
	}
	if string(pass1) != string(pass2) {
		return fmt.Errorf("passphrases do not match")
	}

	if err := secure.SaveUnidocKey(bytesTrimSpace(key), bytesTrimSpace(pass1)); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Saved UNIDOC key.")
	return nil
}

func runConfig(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config subcommand is required")
	}
	switch args[0] {
	case "set":
		return configSet(args[1:])
	default:
		return fmt.Errorf("unknown config subcommand: %s", args[0])
	}
}

func configSet(args []string) error {
	fs := flag.NewFlagSet("config set", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})

	var cfg config.File
	fs.StringVar(&cfg.BaseURL, "base-url", "", "OpenAI compatible base URL")
	fs.StringVar(&cfg.Model, "model", "", "model name")
	fs.StringVar(&cfg.From, "from", "", "source language code")
	fs.StringVar(&cfg.To, "to", "", "target language code")
	fs.StringVar(&cfg.Format, "format", "", "input format default")
	timeout := fs.Duration("timeout", 0, "HTTP timeout (e.g. 120s)")
	maxChars := fs.Int("max-chars", 0, "max chars per translation request")
	endpoint := fs.String("endpoint", "", "endpoint: chat|completion|auto")
	passphraseTTL := fs.Duration("passphrase-ttl", 0, "cache passphrase for duration")

	if err := fs.Parse(args); err != nil {
		return err
	}

	current, err := config.Load()
	if err != nil {
		return err
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "base-url":
			current.BaseURL = cfg.BaseURL
		case "model":
			current.Model = cfg.Model
		case "from":
			current.From = cfg.From
		case "to":
			current.To = cfg.To
		case "format":
			current.Format = cfg.Format
		case "timeout":
			current.TimeoutSeconds = int(timeout.Seconds())
		case "max-chars":
			current.MaxChars = *maxChars
		case "endpoint":
			current.Endpoint = *endpoint
		case "passphrase-ttl":
			current.PassphraseTTLSeconds = int(passphraseTTL.Seconds())
		}
	})

	if err := config.Save(current); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Saved config.")
	return nil
}

func promptHidden(label string) ([]byte, error) {
	fmt.Fprint(os.Stderr, label)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(os.Stderr)
	return pw, nil
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func bytesTrimSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}
