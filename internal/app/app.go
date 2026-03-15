package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wack-backend/internal/config"
	"wack-backend/internal/database"
	"wack-backend/internal/httpserver"
)

type App struct {
	cfg    config.Config
	db     *gorm.DB
	router *gin.Engine
}

func New() (*App, error) {
	cfg := config.Load()

	db, err := database.OpenAndMigrate(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	router, err := httpserver.NewRouter(cfg, db)
	if err != nil {
		return nil, fmt.Errorf("new router: %w", err)
	}

	return &App{
		cfg:    cfg,
		db:     db,
		router: router,
	}, nil
}

func (a *App) Run() error {
	return a.router.Run(":" + a.cfg.Port)
}
