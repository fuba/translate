package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	keyLen       = 32
	saltLen      = 16
	nonceLen     = 12
	scryptN      = 1 << 15
	scryptR      = 8
	scryptP      = 1
	formatVer    = 1
	kdfScrypt    = "scrypt"
	cipherAESGCM = "aes-256-gcm"
)

type envelope struct {
	Version    int    `json:"version"`
	KDF        string `json:"kdf"`
	Cipher     string `json:"cipher"`
	Salt       string `json:"salt"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

func Encrypt(passphrase, plaintext []byte) ([]byte, error) {
	if len(passphrase) == 0 {
		return nil, errors.New("passphrase is required")
	}

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("read salt: %w", err)
	}

	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("read nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	env := envelope{
		Version:    formatVer,
		KDF:        kdfScrypt,
		Cipher:     cipherAESGCM,
		Salt:       base64.RawStdEncoding.EncodeToString(salt),
		Nonce:      base64.RawStdEncoding.EncodeToString(nonce),
		Ciphertext: base64.RawStdEncoding.EncodeToString(ciphertext),
	}

	out, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal envelope: %w", err)
	}
	return out, nil
}

func Decrypt(passphrase, data []byte) ([]byte, error) {
	if len(passphrase) == 0 {
		return nil, errors.New("passphrase is required")
	}

	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("decode envelope: %w", err)
	}
	if env.Version != formatVer {
		return nil, fmt.Errorf("unsupported format version: %d", env.Version)
	}
	if env.KDF != kdfScrypt || env.Cipher != cipherAESGCM {
		return nil, errors.New("unsupported encryption parameters")
	}

	salt, err := base64.RawStdEncoding.DecodeString(env.Salt)
	if err != nil {
		return nil, fmt.Errorf("decode salt: %w", err)
	}
	nonce, err := base64.RawStdEncoding.DecodeString(env.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}
	ciphertext, err := base64.RawStdEncoding.DecodeString(env.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

func deriveKey(passphrase, salt []byte) ([]byte, error) {
	key, err := scrypt.Key(passphrase, salt, scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}
	return key, nil
}
