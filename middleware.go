package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"
)

type authedHTTPHandler func(http.ResponseWriter, *http.Request, *Account) error

func (s *APIServer) authMiddleware(next authedHTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		id, err := validateJWT(token)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Invalid token"})
			//WriteJSON(w, http.StatusUnauthorized, err.Error())
			return
		}

		account, err := s.store.GetAccountByID(id)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "Access forbidden"})
			return
		}

		if err := next(w, r, account); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

// validate the jwt and return account id
func validateJWT(rawToken string) (int, error) {
	// Reject requests that do not include a token.
	if rawToken == "" {
		return 0, errors.New("missing token")
	}

	// Read the server-side secret used to verify the token signature.
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return 0, errors.New("JWT_SECRET is not configured")
	}

	// Parse the token and provide the secret needed to verify its signature.
	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		// Only accept HMAC signing methods; this prevents algorithm substitution.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %q", token.Method.Alg())
		}
		return []byte(secret), nil
	})
	// Reject malformed, incorrectly signed, or expired/otherwise invalid tokens.
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	// Read the token claims so the account identity can be determined.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	// The subject claim identifies the account associated with this token.
	accountID, ok := claims["sub"].(string)
	if !ok {
		return 0, errors.New("missing account id")
	}

	// Convert the subject to a positive account ID for storage lookup.
	id, err := strconv.Atoi(accountID)
	if err != nil || id < 1 {
		return 0, errors.New("invalid account id")
	}
	return id, nil
}
