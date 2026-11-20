package database

import (
	"fmt"
	"strings"
	"time"

	"app/internal/config"
	"app/internal/models"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect membuka koneksi database (mendukung sqlite dan PostgreSQL).
func Connect(cfg *config.Config) (*gorm.DB, error) {
	if strings.EqualFold(cfg.DBConnection, "sqlite") {
		return ConnectSQLite()
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUsername, cfg.DBPassword, cfg.DBDatabase, cfg.DBPort, cfg.DBSSLMode)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
}

// ConnectSQLite membuka SQLite in-memory untuk unit test.
func ConnectSQLite() (*gorm.DB, error) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Post{})
}

func seedUser(db *gorm.DB, username, email, password, role string) error {
	var u models.User
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err := db.Where("username = ?", username).First(&u).Error; err == nil {
		u.Role = role
		u.Password = string(hash)
		return db.Save(&u).Error
	}
	return db.Create(&models.User{Username: username, Email: email, Password: string(hash), Role: role,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}).Error
}

// Seed membuat/memastikan akun admin dan member bawaan memiliki role yang benar.
func Seed(db *gorm.DB, cfg *config.Config) error {
	if err := seedUser(db, cfg.AdminUsername, cfg.AdminUsername+"@delcom.org", cfg.AdminPassword, "admin"); err != nil {
		return err
	}
	return seedUser(db, "member", "member@delcom.org", "member", "member")
}
