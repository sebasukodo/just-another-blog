package handler

import (
	"log/slog"

	"github.com/sebasukodo/just-another-blog/backend/internal/auth"
	"github.com/sebasukodo/just-another-blog/backend/internal/config"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type Handler struct {
	Config    *config.Env
	Auth      *auth.Service
	DbQueries *database.Queries
	Logger    *slog.Logger
}

func NewHandler(db *database.Queries, logger *slog.Logger, cfg *config.Env, auth *auth.Service) *Handler {
	return &Handler{
		Config:    cfg,
		Auth:      auth,
		DbQueries: db,
		Logger:    logger,
	}
}
