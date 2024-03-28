package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/jackc/pgconn"
)

func (pp *postgresProvider) SaveProduct(ctx context.Context, product models.Product, catCodes []int) (string, error) {
	var id int
	err := pp.dbConn.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO "%s" (name, description, category_ids)
VALUES($1,$2,$3)
RETURNING product_id;`,
	pp.cfg.ProductTable),
	product.Name,
	product.Description,
	catCodes).Scan(&id)
	if err == nil {
		return strconv.Itoa(id), nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return "", ErrProductExist
		}
	}
	return "", ErrQuery
}

func (pp *postgresProvider) GetProductById(ctx context.Context, prodId string) (models.Product, error) {
	row := pp.dbConn.QueryRow(ctx, fmt.Sprintf(`
SELECT p.name, p.description, array_agg(c.code) as category_codes
FROM "%s" as p
LEFT JOIN "%s" as c ON c.category_id = ANY (p.category_ids)
WHERE p.product_id = $1
GROUP BY p.product_id`,
	pp.cfg.ProductTable,
	pp.cfg.CatogoryTable),
	prodId)
	var (
		product models.Product
	)
	err := row.Scan(&product.Name, &product.Description, &product.CategoryСodes)
	if err != nil {
		return models.Product{}, ErrQuery
	}
	return product, nil
}

func (pp *postgresProvider) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	rows, err := pp.dbConn.Query(ctx, fmt.Sprintf(`
SELECT p.name, p.description, array_agg(c.code) as category_codes
FROM "%s" as p
LEFT JOIN "%s" as c ON c.category_id = ANY (p.category_ids)
GROUP BY p.product_id`,
	pp.cfg.ProductTable,
	pp.cfg.CatogoryTable))
	if err != nil {
		return nil, ErrQuery
	}
	var outProducts []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.Name, &product.Description, &product.CategoryСodes)
		if err != nil {
			return nil, ErrQuery
		}
		outProducts = append(outProducts, product)
	}
	return outProducts, nil
}

func (pp *postgresProvider) GetProductsByCategory(ctx context.Context, catCode string) ([]models.Product, error) {
	rows, err := pp.dbConn.Query(ctx, fmt.Sprintf(`
SELECT p.name, p.description, array_agg(c.code) as category_codes
FROM "%s" as p
LEFT JOIN "%s" as c ON c.category_id = ANY (p.category_ids)
GROUP BY p.product_id
HAVING $1 = ANY (array_agg(c.code));`,
	pp.cfg.ProductTable,
	pp.cfg.CatogoryTable),
	catCode)
	if err != nil {
		return nil, ErrQuery
	}
	var outProducts []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.Name, &product.Description, &product.CategoryСodes)
		if err != nil {
			return nil, ErrQuery
		}
	}
	return outProducts, nil
}

func (pp *postgresProvider) UpdateProductById(ctx context.Context, id string, newPorductdata models.ProductForPatch, catIds []int) error {
	preparedQuery := fmt.Sprintf("UPDATE \"%s\" SET ", pp.cfg.ProductTable)
	usedFields := 0
	usedData := make([]interface{}, 0)
	if newPorductdata.Name != nil {
		preparedQuery += fmt.Sprintf("\"name\" = $%d, ", usedFields+1)
		usedFields++
		usedData = append(usedData, *newPorductdata.Name)
	}
	if newPorductdata.Description != nil {
		preparedQuery += fmt.Sprintf("\"description\" = $%d, ", usedFields+1)
		usedFields++
		usedData = append(usedData, *newPorductdata.Description)
	}
	if catIds != nil {
		preparedQuery += fmt.Sprintf("\"category_ids\" = $%d, ", usedFields+1)
		usedFields++
		usedData = append(usedData, catIds)
	}
	preparedQuery = preparedQuery[:len(preparedQuery)-2]
	usedData = append(usedData, id)
	_, err := pp.dbConn.Exec(ctx, fmt.Sprintf("%s WHERE \"product_id\" = $%d", preparedQuery, usedFields+1), usedData...)
	if err != nil {
		return ErrQuery
	}
	return nil
}

func (pp *postgresProvider) DeleteProductById(ctx context.Context, id string) error {
	_, err := pp.dbConn.Exec(ctx, fmt.Sprintf("DELETE FROM \"%s\" WHERE \"product_id\" = $1", pp.cfg.ProductTable), id)
	if err != nil {
		return ErrQuery
	}
	return nil
}
