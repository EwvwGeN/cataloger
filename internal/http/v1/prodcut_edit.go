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

type productEditor interface {
	EditProduct(context.Context, string, models.ProductForPatch) (error)
}

func ProductEdit(logger *slog.Logger, validCfg config.Validator, productEditor productEditor) http.HandlerFunc {
	log := logger.With(slog.String("handler", "product_edit"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to edit product")
		prodId, ok := mux.Vars(r)["productId"]
		if !ok || prodId == "" {
			log.Warn("failed to get product id")
			http.Error(w, "error while editing product: empty product id", http.StatusBadRequest)
			return
		}
		req := httpmodels.ProductEditRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "error while decoding request", http.StatusBadRequest)
			return
		}
		log.Debug("got data from request", slog.Any("request_body", req))
		if req.ProductNewData.Name == nil && req.ProductNewData.Description == nil && req.ProductNewData.CategoryСodes == nil {
			log.Warn("nothing to update")
			http.Error(w, "error while editing: nothing to update", http.StatusBadRequest)
			return
		}
		if req.ProductNewData.Name != nil && !validator.ValideteByRegex(*req.ProductNewData.Name, validCfg.ProductNameValidate) {
			log.Info("validate error: incorrect product new name", slog.String("name", *req.ProductNewData.Name))
			http.Error(w, "error while validating product name", http.StatusBadRequest)
			return
		}
		if req.ProductNewData.Description != nil && !validator.ValideteByRegex(*req.ProductNewData.Description, validCfg.ProductDescValidate) {
			log.Info("validate error: incorrect product new description", slog.String("description", *req.ProductNewData.Description))
			http.Error(w, "error while validating product description", http.StatusBadRequest)
			return
		}
		err := productEditor.EditProduct(context.Background(), prodId, req.ProductNewData)
		if err != nil {
			if errors.Is(err, service.ErrProductExist) {
				log.Error("failed to edit category", slog.String("error", service.ErrProductExist.Error()))
				http.Error(w, "error while edditing category: product with this name already exist", http.StatusBadRequest)
				return
			}
			log.Error("error while editing product", slog.Any("product", req.ProductNewData))
			http.Error(w, "failed to edit product", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}