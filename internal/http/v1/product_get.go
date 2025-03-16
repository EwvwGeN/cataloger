package v1

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/cataloger/internal/domain/httpmodels"
	"github.com/EwvwGeN/cataloger/internal/domain/models"
	"github.com/EwvwGeN/cataloger/internal/service"
	"github.com/gorilla/mux"
)

type productOneGetter interface {
	GetOneProduct(context.Context, string) (models.Product, error)
}

type productAllGetter interface {
	GetAllProduct(context.Context) ([]models.Product, error)
}

type productAllByCatCodeGetter interface {
	GetAllProductsByCategory(context.Context, string) ([]models.Product, error)
}

func ProductGetOne(logger *slog.Logger, productOneGetter productOneGetter) http.HandlerFunc {
	log := logger.With(slog.String("handler", "product_get_one"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to get one product")
		prodId, ok := mux.Vars(r)["productId"]
		if !ok || prodId == "" {
			log.Warn("failed to get product id")
			http.Error(w, "error while editing product: empty product id", http.StatusBadRequest)
			return
		}
		product, err := productOneGetter.GetOneProduct(context.Background(), prodId)
		if err != nil {
			if errors.Is(err, service.ErrProductNotFound) {
				log.Warn("failed to get product with this id", slog.String("error", err.Error()))
				http.Error(w, "product with this id not found", http.StatusBadRequest)
				return
			}
			log.Error("failed to get product", slog.String("error", err.Error()))
			http.Error(w, "error while getting product", http.StatusBadRequest)
			return
		}
		res := &httpmodels.ProductGetOneResponse{
			Product: product,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while getting product", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
	}
}

func ProductGetAll(logger *slog.Logger, productAllGetter productAllGetter) http.HandlerFunc {
	log := logger.With(slog.String("handler", "product_get_all"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to get all products")
		products, err := productAllGetter.GetAllProduct(context.Background())
		if err != nil {
			log.Error("failed to get products", slog.String("error", err.Error()))
			http.Error(w, "error while getting products", http.StatusBadRequest)
			return
		}
		res := &httpmodels.ProductGetAllResponse{
			Products: products,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while getting products", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
	}
}

func ProductGetAllByCategory(logger *slog.Logger, pg productAllByCatCodeGetter) http.HandlerFunc {
	log := logger.With(slog.String("handler", "product_get_all_by_category"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to get all products by category code")
		catCode, ok := mux.Vars(r)["catCode"]
		if !ok || catCode == "" {
			log.Warn("failed to get category code")
			http.Error(w, "error while getting category: empty category code", http.StatusBadRequest)
			return
		}
		products, err := pg.GetAllProductsByCategory(context.Background(), catCode)
		if err != nil {
			log.Error("failed to get products", slog.String("error", err.Error()))
			http.Error(w, "error while getting products", http.StatusBadRequest)
			return
		}
		res := &httpmodels.ProductGetAllResponse{
			Products: products,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while getting products", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
	}
}