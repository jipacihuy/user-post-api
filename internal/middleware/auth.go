package middleware

import (
	"context"
	"net/http"
	"strings"

	"user-post-api/internal/utils"
	"user-post-api/pkg/jwt"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondError(w, http.StatusUnauthorized, "missing token")
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondError(w, http.StatusUnauthorized, "invalid token format")
			return
		}
		userID, err := jwt.VerifyToken(parts[1])
		if err != nil {
			utils.RespondError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
