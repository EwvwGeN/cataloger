package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/storage"
)

type categoryService struct {
	log *slog.Logger
	categoryRepo categoryRepo
}

//go:generate go run github.com/vektra/mockery/v2@v2.40.3 --name=categoryRepo --exported
type categoryRepo interface{
	SaveCategory(ctx context.Context, category models.Category) error
	GetCategoryByCode(ctx context.Context, catCode string) (models.Category, error)
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	UpdateCategoryByCode(ctx context.Context, catCode string, catUpdateData models.CategoryForPatch) (error)
	DeleteCategoryBycode(ctx context.Context, catCode string) (error)
}

func NewCategoryService(log *slog.Logger, categoryRepo categoryRepo) *categoryService {
	return &categoryService{
		log: log.With(slog.String("service", "category")),
		categoryRepo: categoryRepo,
	}
}

func (cs *categoryService) AddCategory(ctx context.Context, category models.Category) (error) {
	cs.log.Info("attempt to add category")
	cs.log.Debug("got category", slog.Any("category", category))
	if err := cs.categoryRepo.SaveCategory(ctx, category); err != nil {
		if errors.Is(err, storage.ErrCategoryExist) {
			cs.log.Error("failed to save category", slog.String("error", ErrCategoryExist.Error()))
			return ErrCategoryExist
		}
		cs.log.Error("failed to save category", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (cs *categoryService) GetOneCategory(ctx context.Context, catCode string) (models.Category, error) {
	cs.log.Info("attempt to get category by code")
	cs.log.Debug("got category code", slog.String("code", catCode))
	category, err := cs.categoryRepo.GetCategoryByCode(ctx, catCode)
	if err != nil {
		cs.log.Error("failed to get category by code", slog.String("error", err.Error()))
		return models.Category{}, err
	}
	return category, nil
}

func (cs *categoryService) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	cs.log.Info("attempt to get category by code")
	categories, err := cs.categoryRepo.GetAllCategories(ctx)
	if err != nil {
		cs.log.Error("failed to get categories", slog.String("error", err.Error()))
		return []models.Category{}, err
	}
	return categories, nil
}

func (cs *categoryService) EditCategory(ctx context.Context, catCode string, catUpdateData models.CategoryForPatch) (error) {
	cs.log.Info("attempt to update category")
	cs.log.Debug("got category data", slog.Any("category", catUpdateData))
	if err := cs.categoryRepo.UpdateCategoryByCode(ctx, catCode, catUpdateData); err != nil {
		cs.log.Error("failed to update category", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (cs *categoryService) DeleteCategory(ctx context.Context, catCode string) (error)  {
	cs.log.Info("attempt to delete category")
	cs.log.Debug("got category code", slog.Any("code", catCode))
	if err := cs.categoryRepo.DeleteCategoryBycode(ctx, catCode); err != nil {
		if errors.Is(err, storage.ErrCategoryUsed) {
			cs.log.Error("failed to save category", slog.String("error", ErrCategoryInUse.Error()))
			return ErrCategoryInUse
		}
		cs.log.Error("failed to delete category", slog.String("error", err.Error()))
		return err
	}
	return nil
}