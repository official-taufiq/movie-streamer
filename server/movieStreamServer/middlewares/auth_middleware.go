package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/utils"
)

type Config struct {
	JwtSecret string
}

func (cfg *Config) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := utils.GetBearerToken(r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := utils.ValidateJwt(token, cfg.JwtSecret)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", claims.RegisteredClaims.Subject)
		ctx = context.WithValue(ctx, "role", claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
