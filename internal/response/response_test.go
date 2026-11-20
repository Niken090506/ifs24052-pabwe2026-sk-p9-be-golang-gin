package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	Success(c, http.StatusOK, "ok", gin.H{"k": "v"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"data"`)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	Fail(c2, http.StatusBadRequest, "bad")
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	Error(c3, http.StatusInternalServerError, "err")
	assert.Equal(t, http.StatusInternalServerError, w3.Code)

	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	AbortFail(c4, http.StatusUnauthorized, "unauth")
	assert.Equal(t, http.StatusUnauthorized, w4.Code)

	w5 := httptest.NewRecorder()
	c5, _ := gin.CreateTestContext(w5)
	AbortError(c5, http.StatusBadGateway, "gateway")
	assert.Equal(t, http.StatusBadGateway, w5.Code)
}
