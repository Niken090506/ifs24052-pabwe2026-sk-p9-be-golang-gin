package auth

import (
	"errors"
	"net/http"

	"app/internal/models"
	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{ Svc *Service }

func userMap(u *models.User) gin.H {
	return gin.H{
		"id":         u.ID,
		"username":   u.Username,
		"email":      u.Email,
		"role":       u.Role,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
		"user":       u,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	u, err := h.Svc.Register(req)
	switch {
	case errors.Is(err, ErrInvalid):
		response.Fail(c, http.StatusBadRequest, "Username, email, dan password wajib diisi dengan benar")
	case errors.Is(err, ErrDuplicate):
		response.Fail(c, http.StatusConflict, "Username atau email sudah terdaftar")
	case err != nil:
		response.Error(c, http.StatusInternalServerError, "Gagal mendaftarkan pengguna")
	default:
		response.Success(c, http.StatusCreated, "Registrasi berhasil", userMap(u))
	}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	tok, u, err := h.Svc.Login(req)
	switch {
	case errors.Is(err, ErrInvalid):
		response.Fail(c, http.StatusBadRequest, "Username dan password wajib diisi")
	case errors.Is(err, ErrCredentials):
		response.Fail(c, http.StatusUnauthorized, "Username atau password salah")
	default:
		response.Success(c, http.StatusOK, "Login berhasil", gin.H{
			"token":    tok,
			"user":     u,
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
			"role":     u.Role,
		})
	}
}

func (h *Handler) Logout(c *gin.Context) {
	response.Success(c, http.StatusOK, "Logout berhasil")
}

func (h *Handler) Profile(c *gin.Context) {
	u, err := h.Svc.Repo.FindByID(c.GetUint("user_id"))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Pengguna tidak ditemukan")
		return
	}
	response.Success(c, http.StatusOK, "Profil berhasil diambil", userMap(u))
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	u, err := h.Svc.Update(c.GetUint("user_id"), req)
	switch {
	case errors.Is(err, ErrInvalid):
		response.Fail(c, http.StatusBadRequest, "Email tidak valid")
	case errors.Is(err, ErrDuplicate):
		response.Fail(c, http.StatusConflict, "Username atau email sudah dipakai")
	case err != nil:
		response.Fail(c, http.StatusNotFound, "Pengguna tidak ditemukan")
	default:
		response.Success(c, http.StatusOK, "Profil berhasil diperbarui", userMap(u))
	}
}
