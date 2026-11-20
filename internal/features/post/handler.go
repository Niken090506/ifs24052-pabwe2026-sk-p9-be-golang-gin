package post

import (
	"errors"
	"net/http"
	"strconv"

	"app/internal/models"
	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{ Svc *Service }

func idParam(c *gin.Context) (uint, bool) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "ID tidak valid")
		return 0, false
	}
	return uint(n), true
}

func toPublicPost(p models.Post) gin.H {
	authorName := "Penulis"
	var author gin.H
	if p.User != nil {
		authorName = p.User.Username
		author = gin.H{
			"id":       p.User.ID,
			"username": p.User.Username,
			"role":     p.User.Role,
		}
	}
	return gin.H{
		"id":               p.ID,
		"title":            p.Title,
		"slug":             p.Slug,
		"category":         p.Category,
		"excerpt":          p.Excerpt,
		"content":          p.Content,
		"cover_image_path": p.CoverImagePath,
		"is_published":     p.IsPublished,
		"published_at":     p.PublishedAt,
		"reading_time":     p.ReadingTime,
		"author":           author,
		"author_name":      authorName,
		"username":         authorName,
		"created_at":       p.CreatedAt,
		"updated_at":       p.UpdatedAt,
	}
}

func toPublicPosts(ps []models.Post) []gin.H {
	out := make([]gin.H, 0, len(ps))
	for _, p := range ps {
		out = append(out, toPublicPost(p))
	}
	return out
}

func (h *Handler) Categories(c *gin.Context) {
	response.Success(c, http.StatusOK, "Kategori berhasil diambil", Categories)
}

func (h *Handler) PublicList(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		q = c.Query("search")
	}
	ps, err := h.Svc.Repo.Public(q, c.Query("category"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", toPublicPosts(decorateAll(ps)))
}

func (h *Handler) PublicDetail(c *gin.Context) {
	p, err := h.Svc.Repo.PublicBySlug(c.Param("slug"))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", toPublicPost(*decorate(p)))
}

func (h *Handler) List(c *gin.Context) {
	ps, err := h.Svc.Repo.ByUser(c.GetUint("user_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", decorateAll(ps))
}

func (h *Handler) Stats(c *gin.Context) {
	st := h.Svc.Stats(c.GetUint("user_id"))
	response.Success(c, http.StatusOK, "Statistik berhasil diambil", gin.H{
		"total":           st.Total,
		"total_posts":     st.TotalPosts,
		"total_articles":  st.TotalArticles,
		"total_published": st.TotalPublished,
		"published":       st.Published,
		"total_draft":     st.TotalDraft,
		"draft":           st.Draft,
		"total_selesai":   st.TotalSelesai,
		"belum_selesai":   st.BelumSelesai,
		"stats":           st,
	})
}

func (h *Handler) Create(c *gin.Context) {
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	p, err := h.Svc.Create(c.GetUint("user_id"), req)
	if err == ErrInvalid {
		response.Fail(c, http.StatusBadRequest, "Judul, konten, dan kategori yang valid wajib diisi")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat artikel")
		return
	}
	response.Success(c, http.StatusCreated, "Artikel berhasil dibuat", p)
}

func (h *Handler) Detail(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	p, err := h.Svc.FindOwned(id, c.GetUint("user_id"))
	if errors.Is(err, ErrForbidden) {
		response.Fail(c, http.StatusForbidden, "Anda tidak memiliki akses ke artikel ini")
		return
	}
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", decorate(p))
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	p, err := h.Svc.Update(id, c.GetUint("user_id"), req)
	switch {
	case errors.Is(err, ErrInvalid):
		response.Fail(c, http.StatusBadRequest, "Judul, konten, dan kategori yang valid wajib diisi")
	case errors.Is(err, ErrForbidden):
		response.Fail(c, http.StatusForbidden, "Anda tidak memiliki akses ke artikel ini")
	case err != nil:
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
	default:
		response.Success(c, http.StatusOK, "Artikel berhasil diperbarui", p)
	}
}

func (h *Handler) Toggle(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	p, err := h.Svc.Toggle(id, c.GetUint("user_id"))
	if errors.Is(err, ErrForbidden) {
		response.Fail(c, http.StatusForbidden, "Anda tidak memiliki akses ke artikel ini")
		return
	}
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
		return
	}
	response.Success(c, http.StatusOK, "Status publikasi berhasil diubah", p)
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	p, err := h.Svc.FindOwned(id, c.GetUint("user_id"))
	if errors.Is(err, ErrForbidden) {
		response.Fail(c, http.StatusForbidden, "Anda tidak memiliki akses ke artikel ini")
		return
	}
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
		return
	}
	_ = h.Svc.Repo.Delete(p)
	response.Success(c, http.StatusOK, "Artikel berhasil dihapus")
}

func (h *Handler) AdminUsers(c *gin.Context) {
	us, err := h.Svc.AdminUsers()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil pengguna")
		return
	}
	response.Success(c, http.StatusOK, "Pengguna berhasil diambil", us)
}

func (h *Handler) AdminPosts(c *gin.Context) {
	ps, err := h.Svc.Repo.AllPosts()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", decorateAll(ps))
}

func (h *Handler) AdminUserPosts(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	ps, _ := h.Svc.Repo.ByUser(id)
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", decorateAll(ps))
}
