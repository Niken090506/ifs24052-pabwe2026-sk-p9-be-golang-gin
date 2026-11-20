package ai

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAIClientSearchAndHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock OpenAI server
	aiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer err":
			w.WriteHeader(http.StatusInternalServerError)
		case "Bearer empty":
			_, _ = w.Write([]byte(`{"choices":[]}`))
		case "Bearer badjson":
			_, _ = w.Write([]byte(`not-json`))
		case "Bearer badreply":
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"not-json"}}]}`))
		default:
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"title\":\"T\",\"category\":\"\",\"excerpt\":\"E\",\"content\":\"C\"}"}}]}`))
		}
	}))
	defer aiSrv.Close()

	// Mock Search Engine server
	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		switch q {
		case "err":
			w.WriteHeader(http.StatusBadGateway)
		case "array":
			_, _ = w.Write([]byte(`[{"title":"A","link":"http://a","content":"c"},{"title":"B","url":"http://b","description":"d"},{"title":"3"},{"title":"4"},{"title":"5"},{"title":"6"}]`))
		case "nested":
			_, _ = w.Write([]byte(`{"data":{"results":[{"title":"N","url":"http://n","snippet":"s"}]}}`))
		case "dataarr":
			_, _ = w.Write([]byte(`{"data":[{"title":"D","url":"http://d","snippet":"s"}]}`))
		default:
			_, _ = w.Write([]byte(`{"results":[{"title":"R","url":"http://r","snippet":"s"}]}`))
		}
	}))
	defer searchSrv.Close()

	c := NewClient(aiSrv.URL, "ok", "gpt")
	reply, err := c.Chat(context.Background(), "sys", "usr")
	assert.NoError(t, err)
	assert.Contains(t, reply, `"title":"T"`)

	_, err = NewClient("://bad-url", "ok", "gpt").Chat(context.Background(), "s", "u")
	assert.Error(t, err)
	_, err = NewClient("http://127.0.0.1:1", "ok", "gpt").Chat(context.Background(), "s", "u")
	assert.Error(t, err)
	_, err = NewClient(aiSrv.URL, "err", "gpt").Chat(context.Background(), "s", "u")
	assert.Error(t, err)
	_, err = NewClient(aiSrv.URL, "empty", "gpt").Chat(context.Background(), "s", "u")
	assert.Error(t, err)
	_, err = NewClient(aiSrv.URL, "badjson", "gpt").Chat(context.Background(), "s", "u")
	assert.Error(t, err)

	sc := NewSearchClient(searchSrv.URL, "key")
	srcs, err := sc.Search(context.Background(), "array")
	assert.NoError(t, err)
	assert.Len(t, srcs, 5)
	assert.Equal(t, "http://a", srcs[0].URL)

	srcs, err = sc.Search(context.Background(), "nested")
	assert.NoError(t, err)
	assert.Len(t, srcs, 1)

	srcs, err = sc.Search(context.Background(), "dataarr")
	assert.NoError(t, err)
	assert.Len(t, srcs, 1)

	_, err = sc.Search(context.Background(), "err")
	assert.Error(t, err)
	_, err = NewSearchClient("://bad-url", "").Search(context.Background(), "q")
	assert.Error(t, err)
	_, err = NewSearchClient("http://127.0.0.1:1", "").Search(context.Background(), "q")
	assert.Error(t, err)

	_, err = parseSources([]byte(`not-json`))
	assert.Error(t, err)
	assert.Equal(t, "plain", extractJSON("plain"))
	assert.Contains(t, extractJSON("```json\n{\"a\":1}\n```"), `"a":1`)

	svc := &Service{Chat: c, Search: sc}
	h := &Handler{Svc: svc}
	r := gin.New()
	r.POST("/ai", h.Generate)

	// Bad JSON & empty topic
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/ai", bytes.NewBufferString("{")))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/ai", bytes.NewBufferString(`{"topic":""}`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Invalid category
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/ai", bytes.NewBufferString(`{"topic":"Go","category":"Invalid"}`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// include_web_search=false, use_search_engine=false, use_web_search=true, default search, search error fallback
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/ai", bytes.NewBufferString(`{"topic":"Go","include_web_search":false}`)))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/ai", bytes.NewBufferString(`{"prompt":"Go","category":"Teknologi","use_search_engine":false}`)))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/ai", bytes.NewBufferString(`{"topic":"Go","use_web_search":true}`)))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/ai", bytes.NewBufferString(`{"topic":"err"}`)))
	assert.Equal(t, http.StatusOK, w.Code)

	// Chat error & invalid JSON reply fallback to generated article (200)
	hErr := &Handler{Svc: &Service{Chat: NewClient(aiSrv.URL, "err", "gpt"), Search: sc}}
	r.POST("/ai-err", hErr.Generate)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/ai-err", bytes.NewBufferString(`{"topic":"err"}`)))
	assert.Equal(t, http.StatusOK, w.Code)

	hBad := &Handler{Svc: &Service{Chat: NewClient(aiSrv.URL, "badreply", "gpt"), Search: sc}}
	r.POST("/ai-bad", hBad.Generate)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/ai-bad", bytes.NewBufferString(`{"topic":"Go"}`)))
	assert.Equal(t, http.StatusOK, w.Code)
}
