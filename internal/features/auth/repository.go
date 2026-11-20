package auth

import (
	"app/internal/models"

	"gorm.io/gorm"
)

type Repository struct{ DB *gorm.DB }

func (r *Repository) Create(u *models.User) error { return r.DB.Create(u).Error }

func (r *Repository) FindByID(id uint) (*models.User, error) {
	var u models.User
	if err := r.DB.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByLogin mencari berdasarkan username atau email.
func (r *Repository) FindByLogin(login string) (*models.User, error) {
	var u models.User
	if err := r.DB.Where("username = ? OR email = ?", login, login).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) Exists(username, email string, exceptID uint) bool {
	var n int64
	r.DB.Model(&models.User{}).Where("(username = ? OR email = ?) AND id <> ?", username, email, exceptID).Count(&n)
	return n > 0
}

func (r *Repository) Save(u *models.User) error { return r.DB.Save(u).Error }
