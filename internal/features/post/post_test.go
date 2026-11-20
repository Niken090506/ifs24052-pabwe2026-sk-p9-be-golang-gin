package post

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"app/internal/database"
	"app/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPostPackage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.ConnectSQLite()
	assert.NoError(t, err)
	assert.NoError(t, database.Migrate(db))
	db.Create(&models.User{Username: "u1", Email: "u1@a.com", Password: "x", Role: "member"})

	assert.Equal(t, "artikel", Slugify("🔥🔥🔥"))
	assert.Equal(t, 1, ReadingTime(""))
	assert.Equal(t, 2, ReadingTime(strings.Repeat("kata ", 250)))

	repo := &Repository{DB: db}
	svc := &Service{Repo: repo}
	h := &Handler{Svc: svc}

	r := gin.New()
	withUser := func(fn gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("user_id", uint(1))
			fn(c)
		}
	}
	r.GET("/categories", h.Categories)
	r.GET("/pub", h.PublicList)
	r.GET("/pub/:slug", h.PublicDetail)
	r.GET("/posts", withUser(h.List))
	r.POST("/posts", withUser(h.Create))
	r.GET("/stats", withUser(h.Stats))
	r.GET("/posts/:id", withUser(h.Detail))
	r.PUT("/posts/:id", withUser(h.Update))
	r.PATCH("/posts/:id/toggle", withUser(h.Toggle))
	r.DELETE("/posts/:id", withUser(h.Delete))
	r.GET("/admin/users", h.AdminUsers)
	r.GET("/admin/posts", h.AdminPosts)
	r.GET("/admin/users/:id/posts", h.AdminUserPosts)

	// Categories
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/categories", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	// Bad JSON create
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/posts", bytes.NewBufferString("{")))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Empty title create & invalid category create
	b, _ := json.Marshal(PostRequest{Title: "", Content: "C"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/posts", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(PostRequest{Title: "T", Content: "C", Category: "Salah"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/posts", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Create published post with cover + duplicate slug
	pub := true
	b, _ = json.Marshal(PostRequest{Title: "Judul 1", Content: "Isi 1", CoverImagePath: "/img.png", IsPublished: &pub})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/posts", bytes.NewReader(b)))
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/posts", bytes.NewReader(b)))
	assert.Equal(t, http.StatusCreated, w.Code)

	// Public list & detail (found & not found)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pub?search=Judul&category=Umum", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pub/judul-1", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pub/tidak-ada", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)

	// List, Stats, Detail, Update, Toggle
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/posts", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/stats", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	_ = repo.Count(0, false)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/posts/abc", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/posts/1", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/posts/99", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/posts/abc", bytes.NewBufferString("{}")))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/posts/1", bytes.NewBufferString("{")))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(PostRequest{Title: "", Content: "x"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/posts/1", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	unpub := false
	b, _ = json.Marshal(PostRequest{Title: "Judul 1", Content: "", IsPublished: &unpub})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/posts/1", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(PostRequest{Title: "Judul 1", Content: "Isi baru", IsPublished: &unpub})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/posts/1", bytes.NewReader(b)))
	assert.Equal(t, http.StatusOK, w.Code)

	b, _ = json.Marshal(PostRequest{Title: "Judul 1", Content: "Isi baru 2", IsPublished: &pub})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/posts/1", bytes.NewReader(b)))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/posts/99", bytes.NewReader(b)))
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/posts/abc/toggle", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/posts/99/toggle", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/posts/1/toggle", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	// Admin endpoints
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/admin/users", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/admin/posts", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/admin/users/abc/posts", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/admin/users/1/posts", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/admin/users/999/posts", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Delete
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/posts/abc", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/posts/1", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/posts/1", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)

	// DB error branches when tables are dropped
	_ = db.Migrator().DropTable("posts", "users")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pub", nil))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/posts", nil))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	b, _ = json.Marshal(PostRequest{Title: "Judul X", Content: "Isi X"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/posts", bytes.NewReader(b)))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/admin/users", nil))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/admin/posts", nil))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
