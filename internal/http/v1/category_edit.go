package v1

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/httpmodels"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/service"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/validator"
	"github.com/gorilla/mux"
)

type categoryEditor interface {
	EditCategory(ctx context.Context, catCode string, category models.CategoryForPatch) (error)
}

func CategoryEdit(logger *slog.Logger, validCfg config.Validator, categoryEditor categoryEditor) http.HandlerFunc {
	log := logger.With(slog.String("handler", "category_edit"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to edit category")
		catCode, ok := mux.Vars(r)["catCode"]
		if !ok {
			log.Warn("failed to get category code")
			http.Error(w, "error while editing category: empty category code", http.StatusBadRequest)
			return
		}
		log.Debug("got category code", slog.String("category_code", catCode))
		req := &httpmodels.CategoryEditRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "error while decoding request", http.StatusBadRequest)
			return
		}
		log.Debug("got data from request", slog.Any("request_body", req))
		if req.CategoryNewData.Code == nil && req.CategoryNewData.Name == nil && req.CategoryNewData.Description == nil {
			log.Warn("nothing to update")
			http.Error(w, "error while editing: nothing to update", http.StatusBadRequest)
			return
		}
		if req.CategoryNewData.Name != nil && !validator.ValideteByRegex(*req.CategoryNewData.Name, validCfg.CategoryNameValidate) {
			log.Info("validate error: incorrect category new name", slog.String("name", *req.CategoryNewData.Name))
			http.Error(w, "error while validating category name", http.StatusBadRequest)
			return
		}
		if req.CategoryNewData.Code != nil && !validator.ValideteByRegex(*req.CategoryNewData.Code, validCfg.CategoryCodeValidate) {
			log.Info("validate error: incorrect category new code", slog.String("code", *req.CategoryNewData.Code))
			http.Error(w, "error while validating category code", http.StatusBadRequest)
			return
		}
		if req.CategoryNewData.Description != nil && !validator.ValideteByRegex(*req.CategoryNewData.Description, validCfg.CategoryDescValidate) {
			log.Info("validate error: incorrect category new description", slog.String("description", *req.CategoryNewData.Description))
			http.Error(w, "error while validating category description", http.StatusBadRequest)
			return
		}
		err := categoryEditor.EditCategory(context.Background(), catCode, req.CategoryNewData)
		if err != nil {
			if errors.Is(err, service.ErrCategoryExist) {
				log.Error("failed to add category", slog.String("error", service.ErrCategoryExist.Error()))
				http.Error(w, "error while adding category: category with this code already exist", http.StatusBadRequest)
				return
			}
			log.Error("failed to edit category", slog.String("error", err.Error()))
			http.Error(w, "error while editing category", http.StatusInternalServerError)
			return
		}
		res := &httpmodels.CategoryEditResponse {
			Edited: true,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while editing category", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
	}
}