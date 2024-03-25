package v1

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/httpmodels"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/service"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/validator"
)

type registrator interface{
	RegisterUser(ctx context.Context, email, password string) (error)
}

func Register(logger *slog.Logger, registrator registrator, validateCfg config.Validator) http.HandlerFunc {
	log := logger.With(slog.String("handler", "register"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("got register request")
		req := &httpmodels.RegisterRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("cant decode request body")
			http.Error(w, "error while decoding request", http.StatusBadRequest)
			return
		}
		log.Debug("got data from request", slog.Any("request_body", req))
		
		if !validator.ValideteByRegex(req.Email, validateCfg.EmailValidate) {
			log.Info("validate error: incorrect email", slog.String("email", req.Email))
			http.Error(w, "error while validating email", http.StatusBadRequest)
			return
		}
		if !validator.ValideteByRegex(req.Password, validateCfg.PasswordValidate) {
			log.Info("validate error: incorrect password", slog.String("password", req.Password))
			http.Error(w, "error while validating password", http.StatusBadRequest)
			return
		}
		err := registrator.RegisterUser(context.Background(), req.Email, req.Password)
		if err != nil{
			if errors.Is(err, service.ErrUserExist) {
				log.Warn("failed to save user", slog.String("error", err.Error()))
				http.Error(w, "error while registration: user already exist", http.StatusBadRequest)
				return
			}
			log.Error("failed to save user", slog.String("error", err.Error()))
			http.Error(w, "error while registration", http.StatusInternalServerError)
			return
		}
		res := &httpmodels.RegisterReqsponse{
			Registered: true,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while registration", http.StatusInternalServerError)
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
	}
}