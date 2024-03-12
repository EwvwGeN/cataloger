package v1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/httpmodels"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/gorilla/mux"
)

type categoryOneGetter interface {
	GetOneCategory(ctx context.Context, catCode string) (models.Category, error)
}

type categoryAllGetter interface {
	GetAllCategories(ctx context.Context) ([]models.Category, error)
}

func CategoryGetOne(logger *slog.Logger, categoryOneGetter categoryOneGetter) http.HandlerFunc {
	log := logger.With(slog.String("handler", "category_get_one"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to get category")
		catCode, ok := mux.Vars(r)["catCode"]
		if !ok {
			log.Warn("failed to get category code")
			http.Error(w, "error while getting category: empty category code", http.StatusBadRequest)
			return
		}
		category, err := categoryOneGetter.GetOneCategory(context.Background(), catCode)
		if err != nil {
			log.Error("failed to get category", slog.String("error", err.Error()))
			http.Error(w, "error while getting category", http.StatusInternalServerError)
			return
		}
		res := &httpmodels.CategoryGetOneResponse{
			Category: category,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while getting category", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
	}
}

func CategoryGetAll(logger *slog.Logger, categoryAllGetter categoryAllGetter) http.HandlerFunc {
	log := logger.With(slog.String("handler", "category_get_all"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to get categories")
		categories, err := categoryAllGetter.GetAllCategories(context.Background())
		if err != nil {
			log.Error("failed to get category", slog.String("error", err.Error()))
			http.Error(w, "error while getting categories", http.StatusInternalServerError)
			return
		}
		res := &httpmodels.CategoryGetAllResponse{
			Categories: categories,
		}
		resData, err := json.Marshal(res)
		if err != nil {
			log.Error("cant encode response", slog.Any("response", res), slog.String("error", err.Error()))
			http.Error(w, "error while getting categories", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
	}
}