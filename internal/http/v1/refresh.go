package v1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/httpmodels"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
)

type refresher interface {
	RefreshToken(ctx context.Context, access, refresh string) (models.TokenPair, error)
}

func Refresh(logger *slog.Logger, refresher refresher) http.HandlerFunc {
	log := logger.With(slog.String("handler", "refresh"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("got refresh request")
		req := &httpmodels.RefreshRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("cant decode request body", slog.String("error", err.Error()))
			http.Error(w, "error while decoidng response object", http.StatusBadRequest)
			return
		}
		log.Debug("got data from request", slog.Any("request_body", req))
		if req.TokenPair.AccessToken == "" {
			log.Warn("empty access token")
			http.Error(w, "empty access token", http.StatusBadRequest)
			return
		}
		if req.TokenPair.RefreshToken == "" {
			log.Warn("empty refresh token")
			http.Error(w, "empty refresh token", http.StatusBadRequest)
			return
		}
		tp, err := refresher.RefreshToken(context.Background(), req.TokenPair.AccessToken, req.TokenPair.RefreshToken)
		if err != nil {
			log.Warn("cant refresh token", slog.String("error", err.Error()))
			// edit error handling from RefreshToken()
			http.Error(w, "error while refreshing token", http.StatusBadRequest)
			return
		}
		res := &httpmodels.RefreshResponse{
			TokenPair: models.TokenPair{
				AccessToken: tp.AccessToken,
				RefreshToken: tp.RefreshToken,
			},
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while refreshing token", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
	}
}