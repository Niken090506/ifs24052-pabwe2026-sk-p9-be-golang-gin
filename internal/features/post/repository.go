package post

import (
	"app/internal/models"

	"gorm.io/gorm"
)

type Repository struct{ DB *gorm.DB }

func (r *Repository) Create(p *models.Post) error { return r.DB.Create(p).Error }
func (r *Repository) Save(p *models.Post) error   { return r.DB.Save(p).Error }

func (r *Repository) SlugExists(slug string, exceptID uint) bool {
	var n int64
	r.DB.Model(&models.Post{}).Where("slug = ? AND id <> ?", slug, exceptID).Count(&n)
	return n > 0
}

func (r *Repository) Public(q, category string) ([]models.Post, error) {
	tx := r.DB.Preload("User").Where("is_published = ?", true)
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("LOWER(title) LIKE LOWER(?) OR LOWER(excerpt) LIKE LOWER(?) OR LOWER(content) LIKE LOWER(?)", like, like, like)
	}
	if category != "" && category != "Semua" {
		tx = tx.Where("category = ?", category)
	}
	var out []models.Post
	return out, tx.Order("published_at DESC, id DESC").Find(&out).Error
}

func (r *Repository) PublicBySlug(slug string) (*models.Post, error) {
	var p models.Post
	if err := r.DB.Preload("User").Where("slug = ? AND is_published = ?", slug, true).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) ByUser(uid uint) ([]models.Post, error) {
	var out []models.Post
	return out, r.DB.Where("user_id = ?", uid).Order("id DESC").Find(&out).Error
}

func (r *Repository) FindByID(id uint) (*models.Post, error) {
	var p models.Post
	if err := r.DB.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Owned(id, uid uint) (*models.Post, error) {
	var p models.Post
	if err := r.DB.Where("id = ? AND user_id = ?", id, uid).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Delete(p *models.Post) error { return r.DB.Delete(p).Error }

func (r *Repository) Count(uid uint, publishedOnly bool) int64 {
	var n int64
	tx := r.DB.Model(&models.Post{})
	if uid > 0 {
		tx = tx.Where("user_id = ?", uid)
	}
	if publishedOnly {
		tx = tx.Where("is_published = ?", true)
	}
	tx.Count(&n)
	return n
}

func (r *Repository) AllPosts() ([]models.Post, error) {
	var out []models.Post
	return out, r.DB.Preload("User").Order("id DESC").Find(&out).Error
}

func (r *Repository) Users() ([]models.User, error) {
	var out []models.User
	return out, r.DB.Order("id").Find(&out).Error
}

func (r *Repository) UserExists(id uint) bool {
	var n int64
	r.DB.Model(&models.User{}).Where("id = ?", id).Count(&n)
	return n > 0
}
