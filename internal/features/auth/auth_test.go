package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"app/internal/config"
	"app/internal/database"
	"app/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthPackage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.ConnectSQLite()
	assert.NoError(t, err)
	assert.NoError(t, database.Migrate(db))

	cfg := config.LoadConfig()
	repo := &Repository{DB: db}
	svc := &Service{Repo: repo, Secret: "secret", Cfg: cfg}
	h := &Handler{Svc: svc}

	r := gin.New()
	r.POST("/reg", h.Register)
	r.POST("/login", h.Login)
	r.POST("/logout", h.Logout)
	r.GET("/prof", func(c *gin.Context) {
		if id := c.Query("uid"); id == "1" {
			c.Set("user_id", uint(1))
		} else {
			c.Set("user_id", uint(999))
		}
		h.Profile(c)
	})
	r.PUT("/prof", func(c *gin.Context) {
		if id := c.Query("uid"); id == "1" {
			c.Set("user_id", uint(1))
		} else {
			c.Set("user_id", uint(999))
		}
		h.UpdateProfile(c)
	})

	// Bad JSON
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewBufferString("{")))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Invalid register (empty, short password <6, bad email)
	b, _ := json.Marshal(RegisterRequest{Username: "", Email: "a@b.c", Password: ""})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(RegisterRequest{Username: "u123", Email: "a@b.c", Password: "12345"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(RegisterRequest{Username: "u123", Email: "bad-email", Password: "password123"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(RegisterRequest{Username: "u123", Email: "a@localhost", Password: "password123"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(RegisterRequest{Username: "u123", Email: "a@.com", Password: "password123"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/login", bytes.NewBufferString("{")))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Invalid login
	b, _ = json.Marshal(LoginRequest{})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/login", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Register valid user 1, user 2, and admin account + duplicate (400)
	b, _ = json.Marshal(RegisterRequest{Username: "user1", Email: "u1@a.com", Password: "password123"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(RegisterRequest{Username: "user2", Email: "u2@a.com", Password: "password123"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusCreated, w.Code)

	b, _ = json.Marshal(RegisterRequest{Username: "admin_test", Email: "adm@a.com", Password: "password123"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusCreated, w.Code)

	// Pre-existing admin with role member gets promoted on login
	hash, _ := bcrypt.GenerateFromPassword([]byte("adminpw"), bcrypt.DefaultCost)
	db.Create(&models.User{Username: "admin", Email: "admin@delcom.org", Password: string(hash), Role: "member"})
	b, _ = json.Marshal(LoginRequest{Username: "admin", Password: "adminpw"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/login", bytes.NewReader(b)))
	assert.Equal(t, http.StatusOK, w.Code)

	// Login wrong password, non-existent user, & login by email
	b, _ = json.Marshal(LoginRequest{Username: "ghost", Password: "password123"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/login", bytes.NewReader(b)))
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	b, _ = json.Marshal(LoginRequest{Email: "u1@a.com", Password: "wrong"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/login", bytes.NewReader(b)))
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	b, _ = json.Marshal(LoginRequest{Email: "u1@a.com", Password: "password123"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/login", bytes.NewReader(b)))
	assert.Equal(t, http.StatusOK, w.Code)

	// Profile found & not found
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/prof?uid=1", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/prof?uid=999", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)

	// UpdateProfile bad JSON, invalid email, duplicate (400), not found, and full update
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/prof?uid=1", bytes.NewBufferString("{")))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(UpdateProfileRequest{Email: "not-an-email"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/prof?uid=1", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(UpdateProfileRequest{Username: "user2"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/prof?uid=1", bytes.NewReader(b)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	b, _ = json.Marshal(UpdateProfileRequest{Username: "u1new", Email: "u1new@a.com", Password: "newpassword"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/prof?uid=1", bytes.NewReader(b)))
	assert.Equal(t, http.StatusOK, w.Code)

	b, _ = json.Marshal(UpdateProfileRequest{Username: "x"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PUT", "/prof?uid=999", bytes.NewReader(b)))
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/logout", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	// DB error branch on Register when table is dropped
	_ = db.Migrator().DropTable("users")
	b, _ = json.Marshal(RegisterRequest{Username: "user3", Email: "u3@a.com", Password: "password123"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/reg", bytes.NewReader(b)))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
