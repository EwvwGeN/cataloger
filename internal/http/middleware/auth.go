package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	myhttp "github.com/EwvwGeN/InHouseAd_assignment/internal/http"
)

type jwtParser interface {
	ParseJwt(token string) (map[string]interface{}, error)
}

func AuthMiddleware(logger *slog.Logger, jwtParser jwtParser, next http.HandlerFunc) http.HandlerFunc {
	log := logger.With(slog.String("middleware", "auth"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to authorize")
		authHeader := r.Header.Get("Authorization")
		log.Debug("got authorization header", slog.String("auth", authHeader))
		bearer := strings.Split(authHeader, " ")
		if len(bearer) < 2 && bearer[0] != "Bearer" {
			log.Warn("wrong authorization header")
			http.Error(w, "wrong authorization header", http.StatusBadRequest)
			return
		}
		claims, err := jwtParser.ParseJwt(bearer[1])
		if err != nil {
			log.Warn("failed to parse jwt")
			http.Error(w, "not valid authorization token", http.StatusBadRequest)
			return
		}
		for key, value := range claims {
			r = r.WithContext(context.WithValue(r.Context(), myhttp.ContextKey(key), value))
		}
		next(w,r)
	}
}