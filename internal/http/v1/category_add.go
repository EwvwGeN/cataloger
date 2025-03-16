package v1

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/cataloger/internal/config"
	"github.com/EwvwGeN/cataloger/internal/domain/httpmodels"
	"github.com/EwvwGeN/cataloger/internal/domain/models"
	"github.com/EwvwGeN/cataloger/internal/service"
	"github.com/EwvwGeN/cataloger/internal/validator"
)

type categoryAdder interface {
	AddCategory(ctx context.Context, category models.Category) (error)
}

func CategoryAdd(logger *slog.Logger, validCfg config.Validator, cacategoryAdder categoryAdder) http.HandlerFunc {
	log := logger.With(slog.String("handler", "category_add"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to add category")
		req := &httpmodels.CategoryAddRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "error while decoding request", http.StatusBadRequest)
			return
		}
		log.Debug("got data from request", slog.Any("request_body", req))
		if !validator.ValideteByRegex(req.Category.Name, validCfg.CategoryNameValidate) {
			log.Info("validate error: incorrect category name", slog.String("name", req.Category.Name))
			http.Error(w, "error while validating category name", http.StatusBadRequest)
			return
		}
		if !validator.ValideteByRegex(req.Category.Code, validCfg.CategoryCodeValidate) {
			log.Info("validate error: incorrect category code", slog.String("code", req.Category.Code))
			http.Error(w, "error while validating category code", http.StatusBadRequest)
			return
		}
		if !validator.ValideteByRegex(req.Category.Description, validCfg.CategoryDescValidate) {
			log.Info("validate error: incorrect category description", slog.String("description", req.Category.Description))
			http.Error(w, "error while validating category description", http.StatusBadRequest)
			return
		}
		err := cacategoryAdder.AddCategory(context.Background(), req.Category)
		if err != nil {
			if errors.Is(err, service.ErrCategoryExist) {
				log.Error("failed to add category", slog.String("error", service.ErrCategoryExist.Error()))
				http.Error(w, "error while adding category: category already exist", http.StatusBadRequest)
				return
			}
			log.Error("failed to add category", slog.String("error", err.Error()))
			http.Error(w, "error while adding category", http.StatusBadRequest)
			return
		}
		res := &httpmodels.CategoryAddResponse {
			Added: true,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while adding category", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
	}
}