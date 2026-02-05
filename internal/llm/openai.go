package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sync"
)

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	timeout    time.Duration
	httpClient *http.Client
	endpoint   string

	mu               sync.Mutex
	resolvedEndpoint string
}

type Option func(*Client)

func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.apiKey = key
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

func NewClient(baseURL, model string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("baseURL is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, errors.New("model is required")
	}

	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		// default timeout can be overridden
		timeout:  120 * time.Second,
		endpoint: "chat",
	}
	for _, opt := range opts {
		opt(c)
	}

	c.httpClient = &http.Client{Timeout: c.timeout}
	return c, nil
}

func (c *Client) Translate(ctx context.Context, text, from, to, format string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}

	switch c.effectiveEndpoint(ctx) {
	case "completion":
		return c.translateCompletion(ctx, text, from, to, format)
	default:
		return c.translateChat(ctx, text, from, to, format)
	}
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type completionRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature,omitempty"`
}

type completionResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func chatCompletionsURL(base string) string {
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/v1") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}

func completionsURL(base string) string {
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/v1") {
		return base + "/completions"
	}
	return base + "/v1/completions"
}

func buildSystemPrompt(from, to, format string) string {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)

	src := from
	if src == "" || strings.EqualFold(src, "auto") {
		src = "auto-detect"
	}

	format = strings.ToLower(strings.TrimSpace(format))
	suffix := "Output only the translated text."
	if format == "markdown" {
		suffix = "Preserve Markdown formatting and output only the translated text."
	}

	return fmt.Sprintf("You are a translation engine. Translate from %s to %s. %s", src, to, suffix)
}

func (c *Client) translateChat(ctx context.Context, text, from, to, format string) (string, error) {
	prompt := buildSystemPrompt(from, to, format)
	payload := chatCompletionRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: text},
		},
		Temperature: 0.2,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatCompletionsURL(c.baseURL), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(c.apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("api error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var decoded chatCompletionResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return "", errors.New("api response has no choices")
	}

	content := decoded.Choices[0].Message.Content
	return strings.TrimSpace(content), nil
}

func (c *Client) translateCompletion(ctx context.Context, text, from, to, format string) (string, error) {
	prompt := buildSystemPrompt(from, to, format) + "\n\n" + text
	payload := completionRequest{
		Model:       c.model,
		Prompt:      prompt,
		Temperature: 0.2,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, completionsURL(c.baseURL), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(c.apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("api error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var decoded completionResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return "", errors.New("api response has no choices")
	}

	return strings.TrimSpace(decoded.Choices[0].Text), nil
}

func (c *Client) effectiveEndpoint(ctx context.Context) string {
	if strings.TrimSpace(c.endpoint) == "" || c.endpoint == "chat" {
		return "chat"
	}
	if c.endpoint == "completion" {
		return "completion"
	}
	if c.endpoint != "auto" {
		return "chat"
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.resolvedEndpoint != "" {
		return c.resolvedEndpoint
	}
	c.resolvedEndpoint = resolveEndpoint(ctx, c.httpClient, c.baseURL, c.model, c.apiKey)
	return c.resolvedEndpoint
}

func resolveEndpoint(ctx context.Context, client *http.Client, baseURL, model, apiKey string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL(baseURL), nil)
	if err != nil {
		return "chat"
	}
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "chat"
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "chat"
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "chat"
	}

	type modelEntry struct {
		Model        string   `json:"model"`
		Capabilities []string `json:"capabilities"`
	}
	var payload struct {
		Models []modelEntry `json:"models"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "chat"
	}
	for _, m := range payload.Models {
		if m.Model != model {
			continue
		}
		if hasCapability(m.Capabilities, "completion") && !hasCapability(m.Capabilities, "chat") {
			return "completion"
		}
	}
	return "chat"
}

func modelsURL(base string) string {
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/v1") {
		return base + "/models"
	}
	return base + "/v1/models"
}

func hasCapability(caps []string, target string) bool {
	for _, c := range caps {
		if c == target {
			return true
		}
	}
	return false
}
