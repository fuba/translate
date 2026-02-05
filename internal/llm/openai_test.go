package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type completionReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type completionResp struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func TestAutoEndpointUsesCompletion(t *testing.T) {
	var gotPath string
	var gotReq completionReq

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"models":[{"model":"gpt-oss-20b-mxfp4.gguf","capabilities":["completion"]}]}`))
			return
		}
		gotPath = r.URL.Path
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&gotReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"text":"OK"}]}`))
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "gpt-oss-20b-mxfp4.gguf", WithEndpoint("auto"))
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	out, err := client.Translate(context.Background(), "hello", "en", "ja", "text")
	if err != nil {
		t.Fatalf("Translate error: %v", err)
	}
	if gotPath != "/v1/completions" {
		t.Fatalf("path = %q, want /v1/completions", gotPath)
	}
	if !strings.Contains(gotReq.Prompt, "hello") {
		t.Fatalf("prompt missing input: %q", gotReq.Prompt)
	}
	if out != "OK" {
		t.Fatalf("out = %q", out)
	}
}
