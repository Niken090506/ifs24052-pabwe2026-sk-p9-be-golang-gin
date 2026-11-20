package auth

import (
	"errors"
	"net/mail"
	"strings"

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
}

func (s *Service) Register(req RegisterRequest) (*models.User, error) {
	req.Username, req.Email = strings.TrimSpace(req.Username), strings.TrimSpace(req.Email)
	if req.Username == "" || req.Password == "" {
		return nil, ErrInvalid
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, ErrInvalid
	}
	if s.Repo.Exists(req.Username, req.Email, 0) {
		return nil, ErrDuplicate
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	u := &models.User{Username: req.Username, Email: req.Email, Password: string(hash), Role: "member"}
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
		if _, err := mail.ParseAddress(req.Email); err != nil {
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
