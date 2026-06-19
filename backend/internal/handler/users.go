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

type RegisterUserRequest struct {
	User RegisterUser `json:"user"`
}

type RegisterUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	User LoginUser `json:"user"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	User UpdateUser `json:"user"`
}

type UpdateUser struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
}

type RespondUser struct {
	User User `json:"user"`
}

type User struct {
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

	userInfo := RegisterUserRequest{}

	if err := decoder.Decode(&userInfo); err != nil {
		h.RespondWithError(w, 400, "invalid request body", err.Error())
		return
	}

	if userInfo.User.Email == "" || userInfo.User.Username == "" {
		h.RespondWithError(w, 400, "empty email or username", "user register attempt failed: no email or username")
		return
	}

	if userInfo.User.Password == "" {
		h.RespondWithError(w, 400, "empty password", "user register attempt failed: no password")
		return
	}

	hashedPW, err := auth.HashPassword(userInfo.User.Password)
	if err != nil {
		h.RespondWithError(w, 500, "could not hash password", fmt.Sprintf("could not hash password: %v", err))
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
		h.RespondWithError(w, 201, "user registered successfully, but could not create session", fmt.Sprintf("User %v registered, but could not create session: %v", user.ID, err))
		return
	}

	respBody := RespondUser{
		User: User{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToStringPointer(user.Bio),
			Image:    nullStringToStringPointer(user.Image),
		},
	}

	h.RespondWithJSON(w, 201, respBody)

}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {

	limitedRequest := http.MaxBytesReader(w, r.Body, 2048)
	decoder := json.NewDecoder(limitedRequest)

	userInfo := LoginUserRequest{}

	if err := decoder.Decode(&userInfo); err != nil {
		h.RespondWithError(w, 401, "invalid request body", err.Error())
		return
	}

	if userInfo.User.Email == "" {
		h.RespondWithError(w, 400, "empty email", "user login attempt failed: no email")
		return
	}

	user, err := h.DbQueries.GetUserByEmail(r.Context(), userInfo.User.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 401, "access denied", fmt.Sprintf("login attempt failed, no user found for email %v", userInfo.User.Email))
		} else {
			h.RespondWithDatabaseError(w, err)
		}
		return
	}

	matching, err := auth.CheckPasswordHash(userInfo.User.Password, user.HashedPassword)
	if err != nil {
		h.RespondWithError(w, 401, "access denied", fmt.Sprintf("login attempt failed for user %v - %v", user.ID, err))
		return
	}

	if !matching {
		h.RespondWithError(w, 401, "access denied", fmt.Sprintf("login attempt failed for user %v - wrong password", user.ID))
		return
	}

	token, err := h.Auth.MakeJWT(user.ID, JWTExpiresIn)
	if err != nil {
		h.RespondWithError(w, 500, "access denied", fmt.Sprintf("user %v logged in successfully, but could not create session: %v", user.ID, err))
		return
	}

	respBody := RespondUser{
		User: User{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToStringPointer(user.Bio),
			Image:    nullStringToStringPointer(user.Image),
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
			h.RespondWithError(w, 401, "access denied", fmt.Sprintf("CurrentUser request failed, no user found for id %v", userID))
		} else {
			h.RespondWithDatabaseError(w, err)
		}
		return
	}

	respBody := RespondUser{
		User: User{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToStringPointer(user.Bio),
			Image:    nullStringToStringPointer(user.Image),
		},
	}

	h.RespondWithJSON(w, 200, respBody)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(contextKeyUserID).(uuid.UUID)
	token := r.Context().Value(contextKeyToken).(string)

	decoder := json.NewDecoder(r.Body)

	userInfo := UpdateUserRequest{}

	if err := decoder.Decode(&userInfo); err != nil {
		h.RespondWithError(w, 401, "invalid request body", err.Error())
		return
	}

	updateInfo := database.UpdateUserByIDParams{
		ID: userID,
	}
	if userInfo.User.Username != "" {
		updateInfo.Username = stringToNullString(userInfo.User.Username)
	}
	if userInfo.User.Email != "" {
		updateInfo.Email = stringToNullString(userInfo.User.Email)
	}
	if userInfo.User.Password != "" {
		hashPW, err := auth.HashPassword(userInfo.User.Password)
		if err != nil {
			h.RespondWithError(w, 500, "could not hash new password", err.Error())
			return
		}
		updateInfo.HashedPassword = stringToNullString(hashPW)
	}

	if userInfo.User.Bio != nil && *userInfo.User.Bio == "" {
		_, err := h.DbQueries.ClearUserBio(r.Context(), userID)
		if err != nil {
			h.RespondWithDatabaseError(w, err)
			return
		}
	} else {
		updateInfo.Bio = pointerStringToNullString(userInfo.User.Bio)
	}

	if userInfo.User.Image != nil && *userInfo.User.Image == "" {
		_, err := h.DbQueries.ClearUserImage(r.Context(), userID)
		if err != nil {
			h.RespondWithDatabaseError(w, err)
			return
		}
	} else {
		updateInfo.Image = pointerStringToNullString(userInfo.User.Image)
	}

	user, err := h.DbQueries.UpdateUserByID(r.Context(), updateInfo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 401, "access denied", fmt.Sprintf("UpdateUser request failed, no user found for id %v", userID))
		} else {
			h.RespondWithDatabaseError(w, err)
		}
		return
	}

	respBody := RespondUser{
		User: User{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToStringPointer(user.Bio),
			Image:    nullStringToStringPointer(user.Image),
		},
	}

	h.RespondWithJSON(w, 200, respBody)

}

// only for api testing
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(contextKeyUserID).(uuid.UUID)

	if err := h.DbQueries.DeleteUserByID(r.Context(), userID); err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	w.WriteHeader(204)

}

func pointerStringToNullString(text *string) sql.NullString {
	if text == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *text, Valid: *text != ""}
}

func stringToNullString(text string) sql.NullString {
	return sql.NullString{
		String: text,
		Valid:  text != "",
	}
}

func nullStringToStringPointer(text sql.NullString) *string {
	if text.Valid {
		return &text.String
	}
	return nil
}
