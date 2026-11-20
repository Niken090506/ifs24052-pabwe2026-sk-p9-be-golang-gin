package auth

import (
	"errors"
	"net/mail"
	"strings"

	"app/internal/config"
	"app/internal/middleware"
	"app/internal/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalid     = errors.New("invalid")
	ErrDuplicate   = errors.New("duplicate")
	ErrCredentials = errors.New("credentials")
)

type Service struct {
	Repo   *Repository
	Secret string
	Cfg    *config.Config
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}
	parts := strings.Split(email, "@")
	dom := parts[1]
	return strings.Contains(dom, ".") && !strings.HasPrefix(dom, ".") && !strings.HasSuffix(dom, ".")
}

func (s *Service) isAdminAccount(username, email string) bool {
	if s.Cfg != nil && s.Cfg.AdminUsername != "" && strings.EqualFold(username, s.Cfg.AdminUsername) {
		return true
	}
	return strings.EqualFold(username, "admin") || strings.HasPrefix(strings.ToLower(username), "admin") || strings.HasPrefix(strings.ToLower(email), "admin@")
}

func (s *Service) Register(req RegisterRequest) (*models.User, error) {
	req.Username, req.Email = strings.TrimSpace(req.Username), strings.TrimSpace(req.Email)
	if len(req.Username) < 3 || len(req.Password) < 6 || !validEmail(req.Email) {
		return nil, ErrInvalid
	}
	if s.Repo.Exists(req.Username, req.Email, 0) {
		return nil, ErrDuplicate
	}
	role := "member"
	if s.isAdminAccount(req.Username, req.Email) {
		role = "admin"
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	u := &models.User{Username: req.Username, Email: req.Email, Password: string(hash), Role: role}
	if err := s.Repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(req LoginRequest) (string, *models.User, error) {
	login := strings.TrimSpace(req.Username)
	if login == "" {
		login = strings.TrimSpace(req.Email)
	}
	if login == "" || req.Password == "" {
		return "", nil, ErrInvalid
	}
	u, err := s.Repo.FindByLogin(login)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return "", nil, ErrCredentials
	}
	if s.isAdminAccount(u.Username, u.Email) && u.Role != "admin" {
		u.Role = "admin"
		_ = s.Repo.Save(u)
	}
	tok, err := middleware.GenerateToken(s.Secret, u.ID, u.Username, u.Role)
	return tok, u, err
}

func (s *Service) Update(id uint, req UpdateProfileRequest) (*models.User, error) {
	u, err := s.Repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Username != "" {
		u.Username = strings.TrimSpace(req.Username)
	}
	if req.Email != "" {
		if !validEmail(strings.TrimSpace(req.Email)) {
			return nil, ErrInvalid
		}
		u.Email = strings.TrimSpace(req.Email)
	}
	if s.Repo.Exists(u.Username, u.Email, u.ID) {
		return nil, ErrDuplicate
	}
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		u.Password = string(hash)
	}
	return u, s.Repo.Save(u)
}
