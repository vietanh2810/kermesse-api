package db

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/config"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository/dao"
)

func OpenPostgres(conf *config.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		conf.Host, conf.Port, conf.User, conf.Password, conf.DB,
	)

	gormLogger := createLogger(conf.LogLevel)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open -> %w", err)
	}

	if err = dao.InitTables(db); err != nil {
		return nil, fmt.Errorf("dao.InitTables -> %w", err)
	}

	return db, nil
}

func createLogger(logLevel string) logger.Interface {
	var l logger.LogLevel

	switch strings.ToLower(logLevel) {
	case "silent":
		l = logger.Silent
	case "error":
		l = logger.Error
	case "warn":
		l = logger.Warn
	case "info":
		l = logger.Info
	default:
		l = logger.Error
	}

	return logger.Default.LogMode(l)
}
