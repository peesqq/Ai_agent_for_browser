package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Message struct {
	Role    string
	Content string
}

type Client interface {
	Chat(ctx context.Context, sys string, msgs []Message) (string, error)
}

type openrouterClient struct {
	httpClient *http.Client
	baseURL    string // e.g. https://openrouter.ai/api/v1
	apiKey     string
	model      string
	timeout    time.Duration
}

func NewOpenRouterClient(model string) Client {
	base := os.Getenv("OPENROUTER_BASE_URL")
	if base == "" {
		base = "https://openrouter.ai/api/v1"
	}
	apiKey := "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" //hardcode
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &openrouterClient{
		httpClient: &http.Client{},
		baseURL:    base,
		apiKey:     apiKey,
		model:      model,
		timeout:    30 * time.Second,
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatReq struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float32       `json:"temperature,omitempty"`
	// при необходимости можно добавить другие поля
}

type chatChoice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	// index, finish_reason и т.д. можно добавить при необходимости
}

type chatResp struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Choices []chatChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
	// usage можно добавить если нужно
}

func (o *openrouterClient) Chat(ctx context.Context, sys string, msgs []Message) (string, error) {
	if o.apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY (or OPENAI_API_KEY) is not set")
	}

	// Соберём messages: system (если есть) + переданные
	var m []chatMessage
	if sys != "" {
		m = append(m, chatMessage{Role: "system", Content: sys})
	}
	for _, mm := range msgs {
		role := mm.Role
		if role == "" {
			role = "user"
		}
		m = append(m, chatMessage{Role: role, Content: mm.Content})
	}

	reqBody := chatReq{
		Model:       o.model,
		Messages:    m,
		Temperature: 0.2,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	endpoint := o.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	// опционально: можно добавить заголовки OpenRouter для рейтинга/видимости:
	// httpReq.Header.Set("HTTP-Referer", "<your site>")
	// httpReq.Header.Set("X-Title", "<app name>")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var cr chatResp
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&cr); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	// Если OpenRouter вернул ошибку в body
	if cr.Error != nil && cr.Error.Message != "" {
		return "", fmt.Errorf("api error: %s", cr.Error.Message)
	}

	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("no choices in response, status: %s", resp.Status)
	}

	return cr.Choices[0].Message.Content, nil
}
