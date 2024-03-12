package v1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/httpmodels"
)

type categoryAdder interface {
	AddCategory(ctx context.Context, catName, catCode, catDesc string) (error)
}

func CategoryAdd(logger *slog.Logger, cacategoryAdder categoryAdder) http.HandlerFunc {
	log := logger.With(slog.String("handler", "category_add"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to add category")
		req := &httpmodels.CategoryAddRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "error while decode response object", http.StatusBadRequest)
			return
		}
		log.Debug("got data from request", slog.Any("request_body", req))
		err := cacategoryAdder.AddCategory(context.Background(), req.Category.Name, req.Category.Code, req.Category.Description)
		if err != nil {
			log.Error("failed to add category", slog.String("error", err.Error()))
			http.Error(w, "error while adding category", http.StatusInternalServerError)
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