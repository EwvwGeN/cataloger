package httpmodels

import "github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"

type ProductEditRequest struct {
	ProductNewData models.ProductForPatch `json:"product_new_data"`
}