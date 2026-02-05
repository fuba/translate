package secure

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fuba/translate/internal/config"
	"golang.org/x/term"
)

const passphraseEnv = "TRANSLATE_PASSPHRASE"

func UnidocKeyPath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "unidoc.key"), nil
}

func SaveUnidocKey(key []byte, passphrase []byte) error {
	if len(key) == 0 {
		return errors.New("unidoc key is required")
	}
	path, err := UnidocKeyPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	enc, err := Encrypt(passphrase, key)
	if err != nil {
		return err
	}
	return os.WriteFile(path, enc, 0o600)
}

func LoadUnidocKey(ttl time.Duration) (string, error) {
	if cached, ok := getCachedKey(time.Now()); ok {
		return cached, nil
	}
	if env := strings.TrimSpace(os.Getenv("UNIDOC_LICENSE_API_KEY")); env != "" {
		return env, nil
	}

	path, err := UnidocKeyPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errors.New("unidoc key is not set")
		}
		return "", err
	}

	passphrase := []byte(strings.TrimSpace(os.Getenv(passphraseEnv)))
	if len(passphrase) == 0 {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return "", errors.New("passphrase is required (set TRANSLATE_PASSPHRASE)")
		}
		p, err := promptHidden("Passphrase: ")
		if err != nil {
			return "", err
		}
		passphrase = p
	}

	plain, err := Decrypt(passphrase, data)
	if err != nil {
		return "", err
	}
	key := strings.TrimSpace(string(plain))
	if key == "" {
		return "", errors.New("unidoc key is empty")
	}
	if ttl > 0 {
		setCachedKey(key, time.Now().Add(ttl))
	}
	return key, nil
}

func promptHidden(label string) ([]byte, error) {
	if _, err := fmt.Fprint(os.Stderr, label); err != nil {
		return nil, err
	}
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}
	_, _ = io.WriteString(os.Stderr, "\n")
	return pw, nil
}
