package v1

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/httpmodels"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/service"
)

type loginer interface {
	Login(ctx context.Context, email, password string) (models.TokenPair, error)
}

func Login(logger *slog.Logger, loginer loginer) http.HandlerFunc {
	log := logger.With(slog.String("handler", "login"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("got login request")
		req := &httpmodels.LoginRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("cant decode request body", slog.String("error", err.Error()))
			http.Error(w, "error while decoding request", http.StatusBadRequest)
			return
		}
		log.Debug("got data from request", slog.Any("request_body", req))

		tp, err := loginer.Login(context.Background(), req.Email, req.Password)
		if err != nil {
			log.Warn("cant login user", slog.String("error", err.Error()))
			if errors.Is(err, service.ErrInvalidCredentials) {
				http.Error(w, "error while logging", http.StatusBadRequest)
				return
			}
			http.Error(w, "error while logging", http.StatusInternalServerError)
			return
		}
		res := &httpmodels.LoginResponse{
			TokenPair: models.TokenPair{
				AccessToken: tp.AccessToken,
				RefreshToken: tp.RefreshToken,
			},
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while loggining", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
	}
}