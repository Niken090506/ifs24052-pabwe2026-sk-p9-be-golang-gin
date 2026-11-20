package docs

import (
	"encoding/json"
	"strings"
)

const baseSpec = `{
 "swagger": "2.0",
 "info": {"title": "Delcom Blog API", "description": "REST API Blog Multi-Tier & AI Generator dengan Search Engine", "version": "1.0"},
 "host": "", "schemes": ["http"], "basePath": "/",
 "securityDefinitions": {"BearerAuth": {"type": "apiKey", "name": "Authorization", "in": "header"}},
 "paths": {}
}`

type op struct {
	m, p, desc string
	auth       bool
}

var ops = []op{
	{"get", "/api/public/categories", "Daftar kategori", false},
	{"get", "/api/public/posts", "Daftar artikel terbit (?q, ?category)", false},
	{"get", "/api/public/posts/{slug}", "Detail artikel publik", false},
	{"post", "/api/auth/register", "Registrasi", false},
	{"post", "/api/auth/login", "Login", false},
	{"post", "/api/auth/logout", "Logout", true},
	{"get", "/api/auth/profile", "Profil", true},
	{"put", "/api/auth/profile", "Ubah profil", true},
	{"get", "/api/posts", "Artikel milik penulis", true},
	{"post", "/api/posts", "Buat artikel", true},
	{"get", "/api/posts/stats", "Statistik penulis", true},
	{"post", "/api/posts/ai-generate", "Generate artikel via AI + Search Engine", true},
	{"get", "/api/posts/{id}", "Detail artikel", true},
	{"put", "/api/posts/{id}", "Ubah artikel", true},
	{"patch", "/api/posts/{id}/toggle-publish", "Toggle publikasi", true},
	{"delete", "/api/posts/{id}", "Hapus artikel", true},
	{"get", "/api/admin/users", "Admin: daftar pengguna", true},
	{"get", "/api/admin/posts", "Admin: semua artikel", true},
	{"get", "/api/admin/users/{id}/posts", "Admin: artikel pengguna", true},
}

// GetSwaggerJSON menghasilkan spesifikasi OpenAPI dengan host & scheme dinamis.
func GetSwaggerJSON(host, scheme string) string {
	var spec map[string]interface{}
	_ = json.Unmarshal([]byte(baseSpec), &spec)
	spec["host"], spec["schemes"] = host, []string{scheme}
	paths := spec["paths"].(map[string]interface{})
	for _, o := range ops {
		item, _ := paths[o.p].(map[string]interface{})
		if item == nil {
			item = map[string]interface{}{}
			paths[o.p] = item
		}
		entry := map[string]interface{}{
			"summary":   o.desc,
			"produces":  []string{"application/json"},
			"responses": map[string]interface{}{"200": map[string]string{"description": "OK"}},
		}
		if strings.Contains(o.p, "{") {
			name := "id"
			if strings.Contains(o.p, "{slug}") {
				name = "slug"
			}
			entry["parameters"] = []map[string]interface{}{{"name": name, "in": "path", "required": true, "type": "string"}}
		}
		if o.auth {
			entry["security"] = []map[string][]string{{"BearerAuth": {}}}
		}
		item[o.m] = entry
	}
	out, _ := json.MarshalIndent(spec, "", "  ")
	return string(out)
}
