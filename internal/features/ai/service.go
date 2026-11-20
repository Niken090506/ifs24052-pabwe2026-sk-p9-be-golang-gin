package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalid = errors.New("invalid")

type Chatter interface {
	Chat(ctx context.Context, system, user string) (string, error)
}

type Searcher interface {
	Search(ctx context.Context, query string) ([]Source, error)
}

type Service struct {
	Chat   Chatter
	Search Searcher
}

const systemPrompt = `Kamu adalah penulis blog profesional berbahasa Indonesia. Balas HANYA dengan JSON valid tanpa teks lain, dengan bentuk:
{"title": "...", "category": "...", "excerpt": "ringkasan 1-2 kalimat", "content": "artikel lengkap dalam Markdown"}
Kategori harus salah satu dari: Teknologi, Pendidikan, Gaya Hidup, Bisnis, Opini, Umum.
Jika diberikan sumber web, rujuk sumber tersebut dalam artikel dan sertakan bagian "Sumber" di akhir konten.`

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i, j := strings.Index(s, "{"), strings.LastIndex(s, "}"); i >= 0 && j > i {
		return s[i : j+1]
	}
	return s
}

func shouldSearch(req GenerateRequest) bool {
	if req.IncludeWebSearch != nil {
		return *req.IncludeWebSearch
	}
	if req.UseSearchEngine != nil {
		return *req.UseSearchEngine
	}
	if req.UseWebSearch != nil {
		return *req.UseWebSearch
	}
	return true
}

func (s *Service) Generate(ctx context.Context, req GenerateRequest, validCat func(string) bool) (*Article, error) {
	topic := strings.TrimSpace(req.Topic)
	if topic == "" {
		topic = strings.TrimSpace(req.Prompt)
	}
	if topic == "" {
		return nil, ErrInvalid
	}
	if req.Category != "" && !validCat(req.Category) {
		return nil, ErrInvalid
	}
	sources := []Source{}
	if shouldSearch(req) && s.Search != nil {
		if found, err := s.Search.Search(ctx, topic); err == nil && found != nil {
			sources = found
		}
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Topik artikel: %s\n", topic)
	if req.Category != "" {
		fmt.Fprintf(&sb, "Kategori: %s\n", req.Category)
	}
	if len(sources) > 0 {
		sb.WriteString("\nHasil penelusuran web terkini sebagai konteks:\n")
		for i, src := range sources {
			fmt.Fprintf(&sb, "%d. %s (%s): %s\n", i+1, src.Title, src.URL, src.Snippet)
		}
	}
	reply, err := s.Chat.Chat(ctx, systemPrompt, sb.String())
	if err != nil {
		return nil, err
	}
	var art Article
	if err := json.Unmarshal([]byte(extractJSON(reply)), &art); err != nil {
		return nil, err
	}
	if art.Category == "" || !validCat(art.Category) {
		art.Category = req.Category
		if art.Category == "" {
			art.Category = "Umum"
		}
	}
	art.Sources = sources
	art.SearchSources = sources
	return &art, nil
}
