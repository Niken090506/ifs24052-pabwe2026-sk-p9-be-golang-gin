package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS(), RateLimiter(2))
	r.GET("/ping", Auth("secret"), RequireRole("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// OPTIONS preflight
	reqOpt := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	wOpt := httptest.NewRecorder()
	r.ServeHTTP(wOpt, reqOpt)
	assert.Equal(t, http.StatusNoContent, wOpt.Code)

	// No token
	req1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	// Invalid token
	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req2.Header.Set("Authorization", "Bearer invalid")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// Rate limit exceeded on 3rd request from same IP
	req3 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)

	// Fresh router for role checks
	r2 := gin.New()
	r2.GET("/adm", Auth("secret"), RequireRole("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	tokMember, err := GenerateToken("secret", 1, "m", "member")
	assert.NoError(t, err)
	reqM := httptest.NewRequest(http.MethodGet, "/adm", nil)
	reqM.Header.Set("Authorization", "Bearer "+tokMember)
	wM := httptest.NewRecorder()
	r2.ServeHTTP(wM, reqM)
	assert.Equal(t, http.StatusForbidden, wM.Code)

	tokAdmin, err := GenerateToken("secret", 2, "a", "admin")
	assert.NoError(t, err)
	reqA := httptest.NewRequest(http.MethodGet, "/adm", nil)
	reqA.Header.Set("Authorization", "Bearer "+tokAdmin)
	wA := httptest.NewRecorder()
	r2.ServeHTTP(wA, reqA)
	assert.Equal(t, http.StatusOK, wA.Code)
}
