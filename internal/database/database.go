package database

import (
	"fmt"
	"time"

	"app/internal/config"
	"app/internal/models"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect membuka koneksi PostgreSQL untuk runtime.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUsername, cfg.DBPassword, cfg.DBDatabase, cfg.DBPort, cfg.DBSSLMode)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
}

// ConnectSQLite membuka SQLite in-memory untuk unit test.
func ConnectSQLite() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Post{})
}

func seedUser(db *gorm.DB, username, email, password, role string) error {
	var n int64
	db.Model(&models.User{}).Where("username = ?", username).Count(&n)
	if n > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Create(&models.User{Username: username, Email: email, Password: string(hash), Role: role,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}).Error
}

// Seed membuat akun admin dan member bawaan.
func Seed(db *gorm.DB, cfg *config.Config) error {
	if err := seedUser(db, cfg.AdminUsername, cfg.AdminUsername+"@delcom.org", cfg.AdminPassword, "admin"); err != nil {
		return err
	}
	return seedUser(db, "member", "member@delcom.org", "member", "member")
}
