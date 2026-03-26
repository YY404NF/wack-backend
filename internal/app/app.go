package app

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wack-backend/internal/config"
	"wack-backend/internal/database"
	"wack-backend/internal/httpserver"
	"wack-backend/internal/service"
)

const schoolTimeZone = "Asia/Shanghai"

type App struct {
	cfg    config.Config
	db     *gorm.DB
	router *gin.Engine
	sess   closer
}

type closer interface {
	Close() error
}

func New() (*App, error) {
	cfg := config.Load()
	location, err := time.LoadLocation(schoolTimeZone)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", schoolTimeZone, err)
	}
	time.Local = location

	db, err := database.OpenAndMigrate(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sessionService, err := service.NewSessionService(cfg)
	if err != nil {
		return nil, fmt.Errorf("new session service: %w", err)
	}

	router, err := httpserver.NewRouter(cfg, db, sessionService)
	if err != nil {
		return nil, fmt.Errorf("new router: %w", err)
	}

	return &App{
		cfg:    cfg,
		db:     db,
		router: router,
		sess:   sessionService,
	}, nil
}

func (a *App) Run() error {
	if a.cfg.Host != "" {
		return a.router.Run(a.cfg.Host + ":" + a.cfg.Port)
	}
	return a.router.Run(":" + a.cfg.Port)
}
