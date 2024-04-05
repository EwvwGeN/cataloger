package background

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
)

type productAdder func(ctx context.Context, products []models.Product, categories []models.Category) error

type parser func (config.Validator, []map[string]interface{}) ([]models.Product, []models.Category, error)

func DataCollector(logger *slog.Logger, validCfg config.Validator, parseFunc parser, addFunc productAdder) func(context.Context, string) (error) {
	log := logger.With(slog.String("handler", "data_collector"))
	return func(ctx context.Context, url string) (error) {
		resp, err := http.Get(url)
		if err != nil {
			log.Error("filed to execute get request", slog.String("error", err.Error()))
			return err
		}
		var prodMaps []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&prodMaps)
		if err != nil {
			log.Error("failed to decode response body", slog.String("error", err.Error()))
			return err
		}
		resp.Body.Close()
		log.Debug("got product map", slog.Any("product_map", prodMaps))
		products, categories, err := parseFunc(validCfg, prodMaps)
		if err != nil {
			log.Error("failed to get products from array of maps", slog.String("error", err.Error()))
			return err
		}
		err = addFunc(ctx, products, categories)
		if err != nil {
			log.Error("failed to add products", slog.String("error", err.Error()))
			return err
		}
		return nil
	}
}