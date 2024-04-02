package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"url-shortener/internal/lib/jwt"
	"url-shortener/internal/lib/logger/sl"
)

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrFailedIsAdminCheck = errors.New("failed to check if user is admin")
)

const (
	errorKey   = "error"
	isAdminKey = "isAdminKey"
)

type PermissionProvider interface {
	IsAdmin(ctx context.Context, userId string) (bool, error)
}

// New creates new auth middleware.
func New(
	log *slog.Logger,
	appSecret []byte,
	permProvider PermissionProvider,
) func(next http.Handler) http.Handler {
	const op = "middleware.auth.New"

	log = log.With(slog.String("op", op))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				// It`s ok, if user is not authorized
				next.ServeHTTP(w, r)
				return
			}

			claims, err := jwt.Parse(tokenStr, appSecret)
			if err != nil {
				log.Warn("failed to parse token", sl.Err(err))

				ctx := context.WithValue(r.Context(), errorKey, ErrInvalidToken)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			log.Info("user authorized", slog.Any("claims", claims))

			isAdmin, err := permProvider.IsAdmin(r.Context(), claims.UUID.String())
			if err != nil {
				log.Error("failed to check if user is admin", sl.Err(err))

				ctx := context.WithValue(r.Context(), errorKey, ErrFailedIsAdminCheck)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			ctx := context.WithValue(r.Context(), isAdminKey, isAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken extracts auth token from Authorization header.
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return ""
	}
	return splitToken[1]
}

func ErrorFromContext(ctx context.Context) (error, bool) {
	err, ok := ctx.Value(errorKey).(error)
	return err, ok
}

func IsAdminFromContext(ctx context.Context) (bool, bool) {
	isAdmin, ok := ctx.Value(isAdminKey).(bool)
	return isAdmin, ok
}
