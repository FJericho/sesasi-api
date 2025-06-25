package config

import (
	"fmt"
	"log"
	"time"

	"github.com/FJericho/sesasi-api/internal/entity"
	"github.com/FJericho/sesasi-api/internal/helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type logrusWriter struct {
	Logger *logrus.Logger
}

func (l *logrusWriter) Printf(message string, args ...interface{}) {
	l.Logger.Tracef(message, args...)
}

func NewDatabaseConfig(viper *viper.Viper, log *logrus.Logger) *gorm.DB {
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")
	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	dbName := viper.GetString("database.dbname")
	sslMode := viper.GetString("database.sslmode")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", host, username, password, dbName, port, sslMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(&logrusWriter{Logger: log}, logger.Config{
			SlowThreshold:             time.Second * 5,
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  logger.Warn,
		}),
	})
	if err != nil {
		log.Fatalf("failed to connect database:%+v", err)
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		log.Fatalf("failed to create extension:%+v", err)
	}

	err = db.AutoMigrate(
		&entity.Account{},
		&entity.Permission{},
	)
	if err != nil {
		log.Fatalf("error on running migration : %+v", err)
	}

	SeedAdmin(db, viper)

	return db
}

func SeedAdmin(db *gorm.DB, viper *viper.Viper) {
	hashedPassword, err := helper.HashPassword(viper.GetString("admin.password"))
	if err != nil {
		log.Fatalf("Failed to hash admin password: %+v", err)
	}

	admin := entity.Account{
		Name:     viper.GetString("admin.name"),
		Email:    viper.GetString("admin.email"),
		Password: hashedPassword,
		Role:     entity.ADMIN,
	}

	var existingUser entity.Account
	if err := db.Where("email = ?", admin.Email).First(&existingUser).Error; err != nil {
		db.Create(&admin)
	}
}
