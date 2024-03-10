package v1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/httpmodels"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/validator"
)

type registrator interface{
	RegisterUser(ctx context.Context, email, password string) (string, error)
}

func Register(logger *slog.Logger, registrator registrator, validateCfg config.Validator) http.HandlerFunc {
	log := logger.With(slog.String("handler", "register"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("got register request")
		req := &httpmodels.RegisterRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("cant decode request body")
			http.Error(w, "error while decoidng response object", http.StatusBadRequest)
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
		newID, err := registrator.RegisterUser(context.Background(), req.Email, req.Password)
		if err != nil{
			log.Info("error while registration")
			http.Error(w, "error while registration", http.StatusInternalServerError)
			return
		}
		res := &httpmodels.RegisterReqsponse{
			Registered: true,
			NewUserId: newID,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			// if error happend we still need to send new use id bcz he is already created
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(newID))
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
	}
}