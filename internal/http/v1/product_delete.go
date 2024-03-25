package v1

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type productDeleter interface {
	DelteProduct(context.Context, string) (error)
}

func ProductDelete(logger *slog.Logger, productDeleter productDeleter) http.HandlerFunc {
	log := logger.With(slog.String("handler", "product_delete"))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("attempt to delete product")
		prodId, ok := mux.Vars(r)["productId"]
		if !ok || prodId == "" {
			log.Warn("failed to get product id")
			http.Error(w, "error while editing product: empty product id", http.StatusBadRequest)
			return
		}
		err := productDeleter.DelteProduct(context.Background(), prodId)
		if err != nil {
			log.Error("failed to delete product", slog.String("error", err.Error()))
			http.Error(w, "error while deleting product", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}