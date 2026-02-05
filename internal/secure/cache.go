package secure

import (
	"sync"
	"time"
)

type keyCache struct {
	key     string
	expires time.Time
}

var (
	cacheMu sync.Mutex
	cache   keyCache
)

func getCachedKey(now time.Time) (string, bool) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if cache.key == "" {
		return "", false
	}
	if now.After(cache.expires) {
		cache = keyCache{}
		return "", false
	}
	return cache.key, true
}

func setCachedKey(key string, expires time.Time) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = keyCache{key: key, expires: expires}
}

func clearCachedKey() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = keyCache{}
}
