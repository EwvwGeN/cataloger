package parser

import (
	"fmt"
	"strings"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/validator"
)

// View of data from source
//
// {
//     "name": "grinning face",
//     "category": "smileys and people",
//     "group": "face positive",
//     "htmlCode": [
//       "&#128512;"
//     ],
//     "unicode": [
//       "U+1F600"
//     ]
// },

func ParseFromEmojihub(cfg config.Validator, prodMaps []map[string]interface{}) ([]models.Product, []models.Category, error) {
	if len(prodMaps) == 0 {
		return nil, nil, fmt.Errorf("empry products map")
	}
	products := make([]models.Product, 0, len(prodMaps))
	categories := make([]models.Category, 0, len(prodMaps))
	// it is used to check whether a category with the same name has already been processed.
	// (it would be possible to store categories immediately in the map, but it is inconvenient to process it further)
	categoriesMap := make(map[string]string, len(prodMaps))
	for _, record := range prodMaps {
		// it would be faster to use json.Unmarshal if we had a lot of fields
		pName := record["name"].(string)
		if !validator.ValideteByRegex(pName, cfg.ProductNameValidate) {
			continue
		}
		var product models.Product
		product.Name = pName
		cName := record["category"].(string)
		catCode, ok := categoriesMap[cName]
		if ok {
			product.CategoryСodes = append(product.CategoryСodes, catCode)
		}
		if !ok && validator.ValideteByRegex(cName, cfg.CategoryNameValidate) {
			var category models.Category
			catCode := strings.ReplaceAll(strings.ToLower(cName), " ", "_")
			categoriesMap[cName] = catCode
			category.Name = cName
			category.Code = catCode
			product.CategoryСodes = append(product.CategoryСodes, catCode)
			categories = append(categories, category)
		}
		products = append(products, product)
	}
	if len(products) == 0 {
		return nil, nil, fmt.Errorf("failed to get any valid products")
	}
	return products, categories, nil
}