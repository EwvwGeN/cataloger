package service

import (
	"context"
	"log/slog"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.40.3 --name=productRepo --exported
type productRepo interface {
	SaveProduct(context.Context, models.Product) (string, error)
	GetProductById(context.Context, string) (models.Product, error)
	GetAllProducts(context.Context) ([]models.Product, error)
	UpdateProductById(context.Context, string, models.ProductForPatch) (error)
	DeleteProductById(context.Context, string) (error)
}

type productService struct {
	log *slog.Logger
	productRepo productRepo
}

func NewProductService(logger *slog.Logger, prRepo productRepo) *productService {
	return &productService{
		log: logger.With(slog.String("service", "product")),
		productRepo: prRepo,
	}
}

func (ps *productService) AddProduct(ctx context.Context, product models.Product) (string, error) {
	ps.log.Info("attempt to add product")
	ps.log.Debug("got product", slog.Any("product", product))
	pId, err := ps.productRepo.SaveProduct(ctx, product)
	if err != nil {
		ps.log.Error("failed to save product", slog.String("error", err.Error()))
		return "", err
	}
	return pId, nil
}

func (ps *productService) GetOneProduct(ctx context.Context, prodId string) (models.Product, error) {
	ps.log.Info("attempt to get one product")
	ps.log.Debug("got product id", slog.String("product_id", prodId))
	product, err := ps.productRepo.GetProductById(ctx, prodId)
	if err != nil {
		ps.log.Error("failed to get product by code", slog.String("product_id", prodId), slog.String("error", err.Error()))
		return models.Product{}, err
	}
	return product, nil
}

func (ps *productService) GetAllProduct(ctx context.Context) ([]models.Product, error) {
	ps.log.Info("attempt to get all products")
	products, err := ps.productRepo.GetAllProducts(ctx)
	if err != nil {
		ps.log.Error("failed to get products", slog.String("error", err.Error()))
		return nil, err
	}
	return products, nil
}

func (ps *productService) EditProduct(ctx context.Context, prodId string, prodUpdateData models.ProductForPatch) (error) {
	ps.log.Info("attempt to update product")
	ps.log.Debug("got product data", slog.Any("product", prodUpdateData))
	if err := ps.productRepo.UpdateProductById(ctx, prodId, prodUpdateData); err != nil {
		ps.log.Error("failed to update product", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (ps *productService) DelteProduct(ctx context.Context, prodId string) (error) {
	ps.log.Info("attempt to delete product")
	ps.log.Debug("got product id", slog.Any("product_id", prodId))
	if err := ps.productRepo.DeleteProductById(ctx, prodId); err != nil {
		ps.log.Error("failed to delete product", slog.String("product_id", prodId), slog.String("error", err.Error()))
		return err
	}
	return nil
}


