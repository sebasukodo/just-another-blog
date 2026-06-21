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
	contextKeyUser   contextKey = "user"
)

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerToken, err := h.Auth.GetToken(r.Header)
		if err != nil {
			h.RespondWithError(w, 401, fieldErrorToken, fmt.Sprintf("auth failed - no token: %v", err))
			return
		}

		userID, err := h.Auth.ValidateJWT(headerToken)
		if err != nil {
			h.RespondWithError(w, 401, fieldErrorToken, fmt.Sprintf("auth failed - invalid token: %v", err))
			return
		}

		user, err := h.DbQueries.GetUserByID(r.Context(), userID)
		if err != nil {
			h.RespondWithError(w, 401, fieldErrorToken, fmt.Sprintf("auth failed - user not found: %v", err))
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		ctx = context.WithValue(ctx, contextKeyUser, user)
		ctx = context.WithValue(ctx, contextKeyToken, headerToken)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) OptionalAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerToken, err := h.Auth.GetToken(r.Header)
		if err != nil {
			next(w, r)
			return
		}

		userID, err := h.Auth.ValidateJWT(headerToken)
		if err != nil {
			next(w, r)
			return
		}

		user, err := h.DbQueries.GetUserByID(r.Context(), userID)
		if err != nil {
			next(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		ctx = context.WithValue(ctx, contextKeyUser, user)
		ctx = context.WithValue(ctx, contextKeyToken, headerToken)
		next(w, r.WithContext(ctx))
	}
}
