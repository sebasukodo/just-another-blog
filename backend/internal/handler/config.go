package handler

import (
	"database/sql"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/sebasukodo/just-another-blog/backend/internal/auth"
	"github.com/sebasukodo/just-another-blog/backend/internal/config"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type Handler struct {
	Config    *config.Env
	Auth      *auth.Service
	DbQueries *database.Queries
	Db        *sql.DB
	Logger    *slog.Logger
	Validate  *validator.Validate
}

func NewHandler(rawDb *sql.DB, db *database.Queries, logger *slog.Logger, cfg *config.Env, auth *auth.Service) *Handler {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &Handler{
		Config:    cfg,
		Auth:      auth,
		DbQueries: db,
		Db:        rawDb,
		Logger:    logger,
		Validate:  validate,
	}
}
