package repository

import (
	"github.com/corradoisidoro/sentinel-rbac/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect(databaseUrl string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(databaseUrl), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		panic(err)
	}
}
