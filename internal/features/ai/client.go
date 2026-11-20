package ai

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
)

type Client struct {
	Endpoint, Key, Model string
	HTTP                 *http.Client
}

func NewClient(endpoint, key, model string) *Client {
	return &Client{Endpoint: strings.TrimRight(endpoint, "/"), Key: key, Model: model,
		HTTP: &http.Client{Timeout: 120 * time.Second}}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Chat mengirim prompt ke OpenAI Chat Completions dan mengembalikan isi balasan.
func (c *Client) Chat(ctx context.Context, system, user string) (string, error) {
	payload, _ := json.Marshal(map[string]interface{}{
		"model":    c.Model,
		"messages": []chatMessage{{"system", system}, {"user", user}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Key)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai status %d", resp.StatusCode)
	}
	var out struct {
		Choices []struct {
			Message chatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("openai returned no choices")
	}
	return out.Choices[0].Message.Content, nil
}
