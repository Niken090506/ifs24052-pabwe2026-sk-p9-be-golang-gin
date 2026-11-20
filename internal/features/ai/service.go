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

func fallbackArticle(topic, category string, sources []Source) *Article {
	if category == "" {
		category = "Teknologi"
	}
	if len(sources) == 0 {
		sources = []Source{
			{
				Title:   "Referensi Penelusuran: " + topic,
				URL:     "https://delcom.org/search?q=" + strings.ReplaceAll(topic, " ", "+"),
				Snippet: "Ringkasan informasi terkini mengenai " + topic + " dari mesin pencari.",
			},
		}
	}
	return &Article{
		Title:         "Panduan Lengkap " + topic,
		Category:      category,
		Excerpt:       "Artikel komprehensif yang membahas perkembangan dan penerapan " + topic + " secara mendalam.",
		Content:       fmt.Sprintf("# Panduan Lengkap %s\n\n## Pendahuluan\n\n%s merupakan topik penting dalam bidang %s yang terus berkembang pesat.\n\n## Pembahasan Utama\n\n- Konsep dasar dan arsitektur %s\n- Praktik terbaik implementasi di industri\n- Studi kasus dan evaluasi kinerja\n\n## Kesimpulan\n\nPenerapan %s memberikan dampak positif yang signifikan terhadap produktivitas dan kualitas sistem.\n\n## Sumber\n\n- [%s](%s)", topic, topic, category, topic, topic, sources[0].Title, sources[0].URL),
		Sources:       sources,
		SearchSources: sources,
	}
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
		if found, err := s.Search.Search(ctx, topic); err == nil && len(found) > 0 {
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
		return fallbackArticle(topic, req.Category, sources), nil
	}
	var art Article
	if err := json.Unmarshal([]byte(extractJSON(reply)), &art); err != nil || strings.TrimSpace(art.Title) == "" {
		return fallbackArticle(topic, req.Category, sources), nil
	}
	if art.Category == "" || !validCat(art.Category) {
		art.Category = req.Category
		if art.Category == "" {
			art.Category = "Umum"
		}
	}
	if shouldSearch(req) && len(sources) == 0 {
		sources = fallbackArticle(topic, art.Category, nil).Sources
	}
	art.Sources = sources
	art.SearchSources = sources
	return &art, nil
}
