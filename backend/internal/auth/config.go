package auth

import (
	"log/slog"
)

type Service struct {
	TokenSecret string
	TokenIssuer string
	Logger      *slog.Logger
}

func NewService(secret string, issuer string, logger *slog.Logger) *Service {
	return &Service{
		TokenSecret: secret,
		TokenIssuer: issuer,
		Logger:      logger,
	}
}
