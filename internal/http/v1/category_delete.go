package v1

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type categoryDeleter interface {
	DeleteCategory(ctx context.Context, catCode string) (error)
}

func CategoryDelete(logger *slog.Logger, categoryDeleter categoryDeleter) http.HandlerFunc {
	log := logger.With(slog.String("handler", "category_delete"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to delete category")
		catCode, ok := mux.Vars(r)["catCode"]
		if !ok {
			log.Warn("failed to get category code")
			http.Error(w, "error while editing category: empty category code", http.StatusBadRequest)
			return
		}
		err := categoryDeleter.DeleteCategory(context.Background(), catCode)
		if err != nil {
			log.Error("failed to delete category", slog.String("error", err.Error()))
			http.Error(w, "error while deleting category", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}