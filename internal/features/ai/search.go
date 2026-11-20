package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type SearchClient struct {
	Endpoint string
	Key      string
	HTTP     *http.Client
}

func NewSearchClient(endpoint, key string) *SearchClient {
	return &SearchClient{Endpoint: endpoint, Key: key, HTTP: &http.Client{Timeout: 20 * time.Second}}
}

// Search memanggil Search Engine dan menormalkan hasil ke []Source.
func (s *SearchClient) Search(ctx context.Context, query string) ([]Source, error) {
	u, err := url.Parse(s.Endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if s.Key != "" {
		req.Header.Set("Authorization", "Bearer "+s.Key)
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search engine status %d", resp.StatusCode)
	}
	return parseSources(raw)
}

type rawItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
	Content string `json:"content"`
	Desc    string `json:"description"`
}

func parseSources(raw []byte) ([]Source, error) {
	var env struct {
		Results []rawItem       `json:"results"`
		Items   []rawItem       `json:"items"`
		Data    json.RawMessage `json:"data"`
	}
	var items []rawItem
	if err := json.Unmarshal(raw, &items); err != nil {
		if err := json.Unmarshal(raw, &env); err != nil {
			return nil, err
		}
		items = append(env.Results, env.Items...)
		if len(items) == 0 && len(env.Data) > 0 {
			if json.Unmarshal(env.Data, &items) != nil {
				var d struct {
					Results []rawItem `json:"results"`
				}
				_ = json.Unmarshal(env.Data, &d)
				items = d.Results
			}
		}
	}
	out := make([]Source, 0, len(items))
	for _, it := range items {
		src := Source{Title: it.Title, URL: it.URL, Snippet: it.Snippet}
		if src.URL == "" {
			src.URL = it.Link
		}
		if src.Snippet == "" {
			src.Snippet = it.Content
		}
		if src.Snippet == "" {
			src.Snippet = it.Desc
		}
		out = append(out, src)
	}
	if len(out) > 5 {
		out = out[:5]
	}
	return out, nil
}
