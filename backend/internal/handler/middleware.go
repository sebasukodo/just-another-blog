package handler

import (
	"context"
	"fmt"
	"net/http"
)

type contextKey string

const (
	contextKeyUserID contextKey = "userID"
	contextKeyToken  contextKey = "token"
)

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerToken, err := h.Auth.GetToken(r.Header)
		if err != nil {
			h.RespondWithError(w, 401, "access denied", fmt.Sprintf("auth failed - no token: %v", err))
			return
		}

		userID, err := h.Auth.ValidateJWT(headerToken)
		if err != nil {
			h.RespondWithError(w, 401, "access denied", fmt.Sprintf("auth failed - invalid token: %v", err))
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		ctx = context.WithValue(ctx, contextKeyToken, headerToken)
		next(w, r.WithContext(ctx))
	}
}
