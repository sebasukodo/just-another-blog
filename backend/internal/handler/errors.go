package handler

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Body string `json:"body"`
}

type DBError struct {
	Code       int
	Message    string
	LogMessage string
}

type GenericErrorModel struct {
	Errors map[string][]string `json:"errors"`
}

func validationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "can't be blank"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fieldError.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fieldError.Param())
	default:
		return "is invalid"
	}
}
