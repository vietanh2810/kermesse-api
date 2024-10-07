package app

import (
	"fmt"
	"gorm.io/gorm"
	"os"

	"go.uber.org/zap"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/config"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/db"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/logger"
)

func Start() error {
	conf, err := config.Load("./cmd/app/config.yml")
	if err != nil {
		return fmt.Errorf("failed to initialize config -> %w", err)
	}

	if err = logger.Init(conf.API.Environment); err != nil {
		return fmt.Errorf("failed to initialize logger -> %w", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	var postgresDB *gorm.DB
	if dbURL != "" {
		postgresDB, err = db.OpenPostgresWithURL(dbURL)
	} else {
		postgresDB, err = db.OpenPostgres(conf.Postgres)
	}
	if err != nil {
		return fmt.Errorf("failed to initialize database -> %w", err)
	}

	s := api.NewServer(conf, postgresDB)

	addr := ":" + s.Config.API.Port
	zap.L().Info(fmt.Sprintf("starting server at %v", addr))
	if err = s.Router.Run(addr); err != nil {
		return fmt.Errorf("failed to start the server -> %w", err)
	}

	return nil
}
