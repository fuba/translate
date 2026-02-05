package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestDebugLoggerCalled(t *testing.T) {
	var mu sync.Mutex
	var logs []string
	logger := func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		logs = append(logs, msg)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"models":[{"model":"m","capabilities":["completion"]}]}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"text":"OK"}]}`))
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "m", WithEndpoint("auto"), WithDebugLogger(logger))
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	_, err = client.Translate(context.Background(), "hello", "en", "ja", "text")
	if err != nil {
		t.Fatalf("Translate error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(logs) == 0 {
		t.Fatalf("expected debug logs")
	}
}
