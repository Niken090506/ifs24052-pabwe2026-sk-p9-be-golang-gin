package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSwaggerJSON(t *testing.T) {
	s := GetSwaggerJSON("localhost:8081", "http")
	assert.Contains(t, s, `"host": "localhost:8081"`)
	assert.Contains(t, s, `"/api/public/posts"`)
}
