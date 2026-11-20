package post

import "app/internal/models"

var Categories = []string{"Teknologi", "Pendidikan", "Gaya Hidup", "Bisnis", "Opini", "Umum"}

type PostRequest struct {
	Title          string `json:"title"`
	Category       string `json:"category"`
	Excerpt        string `json:"excerpt"`
	Content        string `json:"content"`
	CoverImagePath string `json:"cover_image_path"`
	IsPublished    *bool  `json:"is_published"`
}

type Stats struct {
	Total          int64 `json:"total"`
	TotalPublished int64 `json:"total_published"`
	TotalDraft     int64 `json:"total_draft"`
	// alias bahasa Indonesia
	TotalSelesai int64 `json:"total_selesai"`
	BelumSelesai int64 `json:"belum_selesai"`
}

type AdminUser struct {
	models.User
	TotalPosts     int64 `json:"total_posts"`
	TotalPublished int64 `json:"total_published"`
}
