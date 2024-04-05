package background

import (
	"context"
	"log/slog"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
)

type productRepo interface {
	SaveProducts(context.Context, []models.Product, [][]int) (error)
}

type categoryRepo interface {
	InserOrGetCategiriesId(context.Context, []models.Category) (map[string]int, error)
}

func GetProductAdder(logger *slog.Logger, prodRepo productRepo, catRepo categoryRepo) func(context.Context, []models.Product, []models.Category) error {
	log := logger.With(slog.String("background_handler", "products_add"))
	return func(ctx context.Context, products []models.Product, categories []models.Category) error {
		catIds, err := catRepo.InserOrGetCategiriesId(ctx, categories)
		if err != nil {
			log.Error("error while getting or inserting categories", slog.String("error", err.Error()))
			return err
		}
		prodCategoriesId := make([][]int, len(products))
		//TODO: rewrite inner loop
		for i, product := range products {
			if len(product.CategoryСodes) == 0 {
				prodCategoriesId[i] = nil
				continue
			}
			for _, category := range product.CategoryСodes {
				if id, ok := catIds[category]; ok {
					prodCategoriesId[i] = append(prodCategoriesId[i], id)
				}
			}
		}
		err = prodRepo.SaveProducts(ctx, products, prodCategoriesId)
		if err != nil {
			log.Error("error while saving products", slog.String("error", err.Error()))
			return err
		}
		return nil
	}
}