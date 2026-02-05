package secure

import (
	"testing"
	"time"
)

func TestCacheKeyExpires(t *testing.T) {
	clearCachedKey()

	now := time.Now()
	setCachedKey("k1", now.Add(10*time.Second))
	if got, ok := getCachedKey(now); !ok || got != "k1" {
		t.Fatalf("expected cached key, got=%q ok=%v", got, ok)
	}

	if got, ok := getCachedKey(now.Add(11 * time.Second)); ok || got != "" {
		t.Fatalf("expected cache expired, got=%q ok=%v", got, ok)
	}
}
