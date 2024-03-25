package v1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/httpmodels"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/validator"
)

type productAdder interface {
	AddProduct(context.Context, models.Product) (string, error)
}

func ProductAdd(logger *slog.Logger, validCfg config.Validator, productAdder productAdder) http.HandlerFunc {
	log := logger.With(slog.String("handler", "product_add"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to add product")
		req := httpmodels.ProductAddRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "error while decoding request", http.StatusBadRequest)
			return
		}
		log.Debug("got data from request", slog.Any("request_body", req))
		if !validator.ValideteByRegex(req.Product.Name, validCfg.ProductNameValidate) {
			log.Info("validate error: incorrect product name", slog.String("name", req.Product.Name))
			http.Error(w, "error while validating product name", http.StatusBadRequest)
			return
		}
		if !validator.ValideteByRegex(req.Product.Description, validCfg.ProductDescValidate) {
			log.Info("validate error: incorrect product description", slog.String("description", req.Product.Description))
			http.Error(w, "error while validating product description", http.StatusBadRequest)
			return
		}
		prodId, err := productAdder.AddProduct(context.Background(), req.Product)
		if err != nil {
			log.Error("error while adding product", slog.Any("product", req.Product))
			http.Error(w, "failed to add product", http.StatusBadRequest)
			return
		}
		res := &httpmodels.ProductAddResponse{
			ProductId: prodId,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while adding product", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
	}
}