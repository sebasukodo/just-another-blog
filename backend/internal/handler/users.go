package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sebasukodo/just-another-blog/backend/internal/auth"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type RegisterRequest struct {
	User RegisterUser `json:"user"`
}

type Respond struct {
	User RespondUser `json:"user"`
}

type RegisterUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
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

	decoder := json.NewDecoder(r.Body)

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
		h.RespondWithError(w, 500, "User could not be registered", fmt.Sprintf("Could not register user %v: %v", userInfo.User.Email, err))
		return
	}

	token, err := h.Auth.MakeJWT(user.ID, JWTExpiresIn)
	if err != nil {
		h.RespondWithError(w, 500, "User registered successfully, but could not create session", fmt.Sprintf("User %v registered, but could not create session: %v", user.ID, err))
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

func nullStringToString(text sql.NullString) *string {
	if text.Valid {
		return &text.String
	}
	return nil
}
