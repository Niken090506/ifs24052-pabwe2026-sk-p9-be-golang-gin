package post

import (
	"net/http"
	"strconv"

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

func (h *Handler) Categories(c *gin.Context) {
	response.Success(c, http.StatusOK, "Kategori berhasil diambil", gin.H{"categories": Categories})
}

func (h *Handler) PublicList(c *gin.Context) {
	ps, err := h.Svc.Repo.Public(c.Query("q"), c.Query("category"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", gin.H{"posts": decorateAll(ps)})
}

func (h *Handler) PublicDetail(c *gin.Context) {
	p, err := h.Svc.Repo.PublicBySlug(c.Param("slug"))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", gin.H{"post": decorate(p)})
}

func (h *Handler) List(c *gin.Context) {
	ps, err := h.Svc.Repo.ByUser(c.GetUint("user_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", gin.H{"posts": decorateAll(ps)})
}

func (h *Handler) Stats(c *gin.Context) {
	response.Success(c, http.StatusOK, "Statistik berhasil diambil", gin.H{"stats": h.Svc.Stats(c.GetUint("user_id"))})
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
	response.Success(c, http.StatusCreated, "Artikel berhasil dibuat", gin.H{"post": p})
}

func (h *Handler) Detail(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	p, err := h.Svc.Repo.Owned(id, c.GetUint("user_id"))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", gin.H{"post": decorate(p)})
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
	if err == ErrInvalid {
		response.Fail(c, http.StatusBadRequest, "Judul, konten, dan kategori yang valid wajib diisi")
		return
	}
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diperbarui", gin.H{"post": p})
}

func (h *Handler) Toggle(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	p, err := h.Svc.Toggle(id, c.GetUint("user_id"))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
		return
	}
	response.Success(c, http.StatusOK, "Status publikasi berhasil diubah", gin.H{"post": p})
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	p, err := h.Svc.Repo.Owned(id, c.GetUint("user_id"))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Artikel tidak ditemukan")
		return
	}
	if err := h.Svc.Repo.Delete(p); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil dihapus")
}

func (h *Handler) AdminUsers(c *gin.Context) {
	us, err := h.Svc.AdminUsers()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil pengguna")
		return
	}
	response.Success(c, http.StatusOK, "Pengguna berhasil diambil", gin.H{"users": us})
}

func (h *Handler) AdminPosts(c *gin.Context) {
	ps, err := h.Svc.Repo.AllPosts()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", gin.H{"posts": decorateAll(ps)})
}

func (h *Handler) AdminUserPosts(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if !h.Svc.Repo.UserExists(id) {
		response.Fail(c, http.StatusNotFound, "Pengguna tidak ditemukan")
		return
	}
	ps, err := h.Svc.Repo.ByUser(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", gin.H{"posts": decorateAll(ps)})
}
