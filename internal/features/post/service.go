package post

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"app/internal/models"
)

var (
	ErrInvalid = errors.New("invalid")
	nonSlug    = regexp.MustCompile(`[^a-z0-9]+`)
)

type Service struct{ Repo *Repository }

func Slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r < unicode.MaxASCII {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	slug := strings.Trim(nonSlug.ReplaceAllString(b.String(), "-"), "-")
	if slug == "" {
		slug = "artikel"
	}
	return slug
}

// ReadingTime = ceil(kata / 200), minimal 1 menit.
func ReadingTime(content string) int {
	words := len(strings.Fields(content))
	m := (words + 199) / 200
	if m < 1 {
		m = 1
	}
	return m
}

func ValidCategory(c string) bool {
	for _, v := range Categories {
		if v == c {
			return true
		}
	}
	return false
}

func (s *Service) uniqueSlug(title string, exceptID uint) string {
	base := Slugify(title)
	slug := base
	for i := 2; s.Repo.SlugExists(slug, exceptID); i++ {
		slug = base + "-" + strconv.Itoa(i)
	}
	return slug
}

func decorate(p *models.Post) *models.Post {
	p.ReadingTime = ReadingTime(p.Content)
	return p
}

func (s *Service) apply(p *models.Post, req PostRequest) error {
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		return ErrInvalid
	}
	if req.Category == "" {
		req.Category = "Umum"
	}
	if !ValidCategory(req.Category) {
		return ErrInvalid
	}
	if p.Title != req.Title || p.Slug == "" {
		p.Slug = s.uniqueSlug(req.Title, p.ID)
	}
	p.Title, p.Category, p.Excerpt, p.Content = strings.TrimSpace(req.Title), req.Category, req.Excerpt, req.Content
	if req.CoverImagePath != "" {
		p.CoverImagePath = req.CoverImagePath
	}
	if req.IsPublished != nil {
		s.setPublish(p, *req.IsPublished)
	}
	return nil
}

func (s *Service) setPublish(p *models.Post, v bool) {
	p.IsPublished = v
	if v && p.PublishedAt == nil {
		now := time.Now()
		p.PublishedAt = &now
	}
	if !v {
		p.PublishedAt = nil
	}
}

func (s *Service) Create(uid uint, req PostRequest) (*models.Post, error) {
	p := &models.Post{UserID: uid}
	if err := s.apply(p, req); err != nil {
		return nil, err
	}
	if err := s.Repo.Create(p); err != nil {
		return nil, err
	}
	return decorate(p), nil
}

func (s *Service) Update(id, uid uint, req PostRequest) (*models.Post, error) {
	p, err := s.Repo.Owned(id, uid)
	if err != nil {
		return nil, err
	}
	if err := s.apply(p, req); err != nil {
		return nil, err
	}
	return decorate(p), s.Repo.Save(p)
}

func (s *Service) Toggle(id, uid uint) (*models.Post, error) {
	p, err := s.Repo.Owned(id, uid)
	if err != nil {
		return nil, err
	}
	s.setPublish(p, !p.IsPublished)
	return decorate(p), s.Repo.Save(p)
}

func (s *Service) Stats(uid uint) Stats {
	total, pub := s.Repo.Count(uid, false), s.Repo.Count(uid, true)
	return Stats{Total: total, TotalPublished: pub, TotalDraft: total - pub, TotalSelesai: pub, BelumSelesai: total - pub}
}

func decorateAll(ps []models.Post) []models.Post {
	for i := range ps {
		decorate(&ps[i])
	}
	return ps
}

func (s *Service) AdminUsers() ([]AdminUser, error) {
	users, err := s.Repo.Users()
	if err != nil {
		return nil, err
	}
	out := make([]AdminUser, 0, len(users))
	for _, u := range users {
		out = append(out, AdminUser{User: u, TotalPosts: s.Repo.Count(u.ID, false), TotalPublished: s.Repo.Count(u.ID, true)})
	}
	return out, nil
}
