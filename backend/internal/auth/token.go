package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func (s *Service) GetToken(headers http.Header) (string, error) {

	authHeader := headers.Get("Authorization")

	if authHeader == "" {
		s.Logger.Warn("authentication failed - no authorization header set")
		return "", fmt.Errorf("access denied")
	}

	prefix := "Token "
	if !strings.HasPrefix(authHeader, prefix) {
		s.Logger.Warn("authentication failed - wrong token prefix")
		return "", fmt.Errorf("access denied")
	}

	authHeader = strings.TrimPrefix(authHeader, prefix)

	if authHeader == "" {
		s.Logger.Warn("authentication failed - no token content")
		return "", fmt.Errorf("access denied")
	}

	return authHeader, nil

}

func (s *Service) MakeJWT(userID uuid.UUID, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()

	registerClaims := jwt.RegisteredClaims{
		Issuer:    s.TokenIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, registerClaims)

	jw, err := token.SignedString([]byte(s.TokenSecret))
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not create JWT: %v", err))
		return "", fmt.Errorf("invalid token")
	}

	return jw, nil
}

func (s *Service) ValidateJWT(token string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.TokenSecret), nil
	}

	_, err := jwt.ParseWithClaims(token, &claims, keyFunc)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not parse jwt token with claims: %v", err))
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	if claims.Issuer != s.TokenIssuer {
		s.Logger.Warn(fmt.Sprintf("authentication failed - wrong Issuer used: %v", claims.Issuer))
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not parse userid of jwt token %v", token))
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	return userID, nil
}
