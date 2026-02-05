package secure

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	pass := []byte("secret-pass")
	plain := []byte("hello")

	enc, err := Encrypt(pass, plain)
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	dec, err := Decrypt(pass, enc)
	if err != nil {
		t.Fatalf("Decrypt error: %v", err)
	}

	if string(dec) != string(plain) {
		t.Fatalf("Decrypt got %q, want %q", string(dec), string(plain))
	}
}
