package ai

import (
	"errors"
	"net/http"

	"app/internal/features/post"
	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{ Svc *Service }

func (h *Handler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	art, err := h.Svc.Generate(c.Request.Context(), req, post.ValidCategory)
	if errors.Is(err, ErrInvalid) {
		response.Fail(c, http.StatusBadRequest, "Topik wajib diisi dan kategori harus valid")
		return
	}
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Gagal menghasilkan artikel dengan AI")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil dihasilkan", gin.H{
		"title": art.Title, "category": art.Category, "excerpt": art.Excerpt,
		"content": art.Content, "sources": art.Sources, "article": art,
	})
}
