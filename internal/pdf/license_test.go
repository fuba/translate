package pdf

import "testing"

func TestIsLicenseAlreadySet(t *testing.T) {
	if isLicenseAlreadySet(nil) {
		t.Fatalf("expected false for nil error")
	}
	if !isLicenseAlreadySet(errString("license key already set")) {
		t.Fatalf("expected true for license set error")
	}
	if isLicenseAlreadySet(errString("other error")) {
		t.Fatalf("expected false for other error")
	}
}

type errString string

func (e errString) Error() string {
	return string(e)
}
