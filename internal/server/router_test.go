package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeChat struct{ err error }

func (f fakeChat) Chat(ctx context.Context, s, u string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "```json\n{\"title\":\"Judul AI\",\"category\":\"Teknologi\",\"excerpt\":\"ex\",\"content\":\"# isi\"}\n```", nil
}

type fakeSearch struct{}

func (fakeSearch) Search(ctx context.Context, q string) ([]ai.Source, error) {
	return []ai.Source{{Title: "t", URL: "http://x", Snippet: "s"}}, nil
}

func setup(t *testing.T, chatErr error) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := config.LoadConfig()
	db, err := database.ConnectSQLite()
	assert.NoError(t, err)
	assert.NoError(t, database.Migrate(db))
	assert.NoError(t, database.Seed(db, cfg))
	return NewRouter(cfg, db, &ai.Service{Chat: fakeChat{chatErr}, Search: fakeSearch{}})
}

func call(r *gin.Engine, method, path, token string, body interface{}) (int, map[string]interface{}) {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func login(t *testing.T, r *gin.Engine, u, p string) string {
	code, out := call(r, "POST", "/api/auth/login", "", map[string]string{"username": u, "password": p})
	assert.Equal(t, 200, code)
	return out["data"].(map[string]interface{})["token"].(string)
}

func TestFullFlow(t *testing.T) {
	r := setup(t, nil)

	code, _ := call(r, "GET", "/", "", nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "GET", "/nope", "", nil)
	assert.Equal(t, 404, code)
	code, _ = call(r, "GET", "/api.json", "", nil)
	assert.Equal(t, 200, code)

	reqTLS := httptest.NewRequest("GET", "/api.json", nil)
	reqTLS.Header.Set("X-Forwarded-Proto", "https")
	wTLS := httptest.NewRecorder()
	r.ServeHTTP(wTLS, reqTLS)
	assert.Equal(t, 200, wTLS.Code)

	// auth
	code, _ = call(r, "POST", "/api/auth/register", "", map[string]string{"username": "budi", "email": "budi@x.com", "password": "password123"})
	assert.Equal(t, 201, code)
	code, _ = call(r, "POST", "/api/auth/register", "", map[string]string{"username": "budi", "email": "budi@x.com", "password": "password123"})
	assert.Equal(t, 400, code)
	code, _ = call(r, "POST", "/api/auth/register", "", map[string]string{"username": "", "email": "bad", "password": ""})
	assert.Equal(t, 400, code)
	code, _ = call(r, "POST", "/api/auth/login", "", map[string]string{"username": "budi", "password": "salah"})
	assert.Equal(t, 401, code)
	tok := login(t, r, "budi", "password123")
	code, _ = call(r, "GET", "/api/auth/profile", "", nil)
	assert.Equal(t, 401, code)
	code, _ = call(r, "GET", "/api/auth/profile", tok, nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "PUT", "/api/auth/profile", tok, map[string]string{"email": "budi2@x.com"})
	assert.Equal(t, 200, code)
	code, _ = call(r, "POST", "/api/auth/logout", tok, nil)
	assert.Equal(t, 200, code)

	// posts
	code, out := call(r, "POST", "/api/posts", tok, map[string]interface{}{"title": "Halo Dunia!", "content": "isi artikel", "category": "Teknologi", "is_published": true})
	assert.Equal(t, 201, code)
	slug := out["data"].(map[string]interface{})["slug"]
	assert.Equal(t, "halo-dunia", slug)
	_, out = call(r, "POST", "/api/posts", tok, map[string]interface{}{"title": "Halo Dunia", "content": "isi"})
	assert.Equal(t, "halo-dunia-2", out["data"].(map[string]interface{})["slug"])

	code, out = call(r, "GET", "/api/public/posts?q=halo&category=Teknologi", "", nil)
	assert.Equal(t, 200, code)
	assert.Len(t, out["data"], 1)
	code, _ = call(r, "GET", "/api/public/posts/halo-dunia", "", nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "GET", "/api/public/posts/halo-dunia-2", "", nil)
	assert.Equal(t, 404, code)

	code, _ = call(r, "PATCH", "/api/posts/2/toggle-publish", tok, nil)
	assert.Equal(t, 200, code)
	code, out = call(r, "GET", "/api/posts/stats", tok, nil)
	assert.Equal(t, float64(2), out["data"].(map[string]interface{})["total_published"])
	code, _ = call(r, "GET", "/api/posts", tok, nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "GET", "/api/posts/1", tok, nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "PUT", "/api/posts/1", tok, map[string]interface{}{"title": "Baru", "content": "x"})
	assert.Equal(t, 200, code)
	code, _ = call(r, "PUT", "/api/posts/1", tok, map[string]interface{}{"title": "", "content": "x"})
	assert.Equal(t, 400, code)
	code, _ = call(r, "GET", "/api/posts/99", tok, nil)
	assert.Equal(t, 404, code)
	code, _ = call(r, "GET", "/api/public/categories", "", nil)
	assert.Equal(t, 200, code)

	// AI
	code, out = call(r, "POST", "/api/posts/ai-generate", tok, map[string]interface{}{"topic": "Go"})
	assert.Equal(t, 200, code)
	assert.Len(t, out["data"].(map[string]interface{})["sources"], 1)
	code, _ = call(r, "POST", "/api/posts/ai-generate", tok, map[string]interface{}{"topic": ""})
	assert.Equal(t, 400, code)

	// admin
	code, _ = call(r, "GET", "/api/admin/users", tok, nil)
	assert.Equal(t, 403, code)
	adm := login(t, r, "admin", "admin")
	code, _ = call(r, "GET", "/api/admin/users", adm, nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "GET", "/api/admin/posts", adm, nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "GET", "/api/admin/users/3/posts", adm, nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "GET", "/api/admin/users/99/posts", adm, nil)
	assert.Equal(t, 200, code)

	code, _ = call(r, "DELETE", "/api/posts/1", tok, nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "DELETE", "/api/posts/1", tok, nil)
	assert.Equal(t, 404, code)
}

func TestAIFallback(t *testing.T) {
	r := setup(t, errors.New("boom"))
	tok := login(t, r, "member", "member")
	code, _ := call(r, "POST", "/api/posts/ai-generate", tok, map[string]interface{}{"topic": "Go"})
	assert.Equal(t, http.StatusOK, code)
}
