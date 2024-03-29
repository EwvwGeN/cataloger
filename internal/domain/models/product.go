package models

type Product struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	CategoryСodes []string `json:"category_codes"`
}

type ProductForPatch struct {
	Name          *string  `json:"name"`
	Description   *string  `json:"description"`
	CategoryСodes []string `json:"category_codes,omitempty"`
}