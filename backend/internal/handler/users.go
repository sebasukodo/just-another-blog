package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sebasukodo/just-another-blog/backend/internal/auth"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type RegisterRequest struct {
	User RegisterUser `json:"user"`
}

type RegisterUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	User LoginUser `json:"user"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Respond struct {
	User RespondUser `json:"user"`
}

type RespondUser struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Token    string  `json:"token"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
}

const JWTExpiresIn = time.Duration(15) * time.Minute

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {

	limitedRequest := http.MaxBytesReader(w, r.Body, 2048)
	decoder := json.NewDecoder(limitedRequest)

	userInfo := RegisterRequest{}

	if err := decoder.Decode(&userInfo); err != nil {
		h.RespondWithError(w, 400, "Invalid Request Body", err.Error())
		return
	}

	hashedPW, err := auth.HashPassword(userInfo.User.Password)
	if err != nil {
		h.RespondWithError(w, 500, "Could not hash password", fmt.Sprintf("Could not hash password: %v", err))
		return
	}

	user, err := h.DbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Username:       userInfo.User.Username,
		Email:          userInfo.User.Email,
		HashedPassword: hashedPW,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	token, err := h.Auth.MakeJWT(user.ID, JWTExpiresIn)
	if err != nil {
		h.RespondWithError(w, 201, "User registered successfully, but could not create session", fmt.Sprintf("User %v registered, but could not create session: %v", user.ID, err))
		return
	}

	respBody := Respond{
		User: RespondUser{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToString(user.Bio),
			Image:    nullStringToString(user.Image),
		},
	}

	h.RespondWithJSON(w, 201, respBody)

}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {

	limitedRequest := http.MaxBytesReader(w, r.Body, 2048)
	decoder := json.NewDecoder(limitedRequest)

	userInfo := LoginRequest{}

	if err := decoder.Decode(&userInfo); err != nil {
		h.RespondWithError(w, 401, "Invalid Request Body", err.Error())
		return
	}

	user, err := h.DbQueries.GetUserByEmail(r.Context(), userInfo.User.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 401, "Access Denied", fmt.Sprintf("Login attempt failed, no user found for email %v", userInfo.User.Email))
		} else {
			h.RespondWithDatabaseError(w, err)
		}
		return
	}

	matching, err := auth.CheckPasswordHash(userInfo.User.Password, user.HashedPassword)
	if err != nil {
		h.RespondWithError(w, 401, "Access Denied", fmt.Sprintf("Login attempt failed for user %v - %v", user.ID, err))
		return
	}

	if !matching {
		h.RespondWithError(w, 401, "Access Denied", fmt.Sprintf("Login attempt failed for user %v - wrong password", user.ID))
		return
	}

	token, err := h.Auth.MakeJWT(user.ID, JWTExpiresIn)
	if err != nil {
		h.RespondWithError(w, 500, "Access Denied", fmt.Sprintf("User %v logged in successfully, but could not create session: %v", user.ID, err))
		return
	}

	respBody := Respond{
		User: RespondUser{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToString(user.Bio),
			Image:    nullStringToString(user.Image),
		},
	}

	h.RespondWithJSON(w, 200, respBody)
}

func (h *Handler) CurrentUser(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(contextKeyUserID).(uuid.UUID)
	token := r.Context().Value(contextKeyToken).(string)

	user, err := h.DbQueries.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 401, "Access Denied", fmt.Sprintf("CurrentUser request failed, no user found for id %v", userID))
		} else {
			h.RespondWithDatabaseError(w, err)
		}
		return
	}

	respBody := Respond{
		User: RespondUser{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToString(user.Bio),
			Image:    nullStringToString(user.Image),
		},
	}

	h.RespondWithJSON(w, 200, respBody)
}

func nullStringToString(text sql.NullString) *string {
	if text.Valid {
		return &text.String
	}
	return nil
}
