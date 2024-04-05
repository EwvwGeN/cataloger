package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

func (pp *postgresProvider) SaveProduct(ctx context.Context, product models.Product, catIds []int) (string, error) {
	transaction, err := pp.dbConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", ErrStartTx
	}
	var id int
	err = transaction.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO "%s" (name, description)
VALUES($1,$2)
RETURNING product_id;`,
	pp.cfg.ProductTable),
	product.Name,
	product.Description).Scan(&id)
	if err != nil {
		if err := transaction.Rollback(ctx); err != nil {
			return "", ErrRollbackTx
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return "", ErrProductExist
			}
		}
		return "", ErrQuery
	}
	for _, сid := range catIds {
		_, err = transaction.Exec(ctx, fmt.Sprintf(`
INSERT INTO "%s" (product_id, category_id)
VALUES ($1, $2)`,
		pp.cfg.ProductCategoryTable),
		id,
		сid)
		if err != nil {
			transaction.Rollback(ctx)
			return "", ErrQuery
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return "", ErrCommitTx
	}
	return strconv.Itoa(id), nil
}

func (pp *postgresProvider) SaveProducts(ctx context.Context, products []models.Product, catsIds [][]int) (error) {
	transaction, err := pp.dbConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ErrStartTx
	}
	for idx, product := range products {
		var prodId int
		err = transaction.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO "%s" (name, description)
VALUES($1,$2)
ON CONFLICT (name) DO NOTHING
RETURNING product_id;`,
			pp.cfg.ProductTable),
			product.Name,
			product.Description).Scan(&prodId)
		if err != nil {
			continue
		}

		for _, сid := range catsIds[idx] {
			_, err = transaction.Exec(ctx, fmt.Sprintf(`
INSERT INTO "%s" (product_id, category_id)
VALUES ($1, $2)
ON CONFLICT (product_id, category_id) DO NOTHING;`,
			pp.cfg.ProductCategoryTable),
			prodId,
			сid)
			if err != nil {
				continue
			}
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return ErrCommitTx
	}
	return nil
}

func (pp *postgresProvider) GetProductById(ctx context.Context, prodId string) (models.Product, error) {
	row := pp.dbConn.QueryRow(ctx, fmt.Sprintf(`
SELECT p.product_id, p.name, p.description, array_agg(c.code) as category_codes
FROM "%s" as p
LEFT JOIN "%s" as pc ON pc.product_id = $1
Left JOIN "%s" as c ON c.category_id = pc.category_id
WHERE p.product_id = $1
GROUP BY p.product_id`,
	pp.cfg.ProductTable,
	pp.cfg.ProductCategoryTable,
	pp.cfg.CatogoryTable),
	prodId)
	var (
		product models.Product
	)
	err := row.Scan(&product.Id, &product.Name, &product.Description, &product.CategoryСodes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Product{}, ErrProductNotFound
		}
		return models.Product{}, ErrQuery
	}
	return product, nil
}

func (pp *postgresProvider) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	rows, err := pp.dbConn.Query(ctx, fmt.Sprintf(`
SELECT p.product_id, p.name, p.description,
CASE 
	WHEN COUNT(pc.category_id) = 0 THEN NULL 
	ELSE array_agg(c.code) 
END as category_codes
FROM "%s" as p
LEFT JOIN "%s" as pc ON pc.product_id = p.product_id
Left JOIN "%s" as c ON c.category_id = pc.category_id
GROUP BY p.product_id`,
	pp.cfg.ProductTable,
	pp.cfg.ProductCategoryTable,
	pp.cfg.CatogoryTable))
	if err != nil {
		return nil, ErrQuery
	}
	var outProducts []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.Id, &product.Name, &product.Description, &product.CategoryСodes)
		if err != nil {
			return nil, ErrQuery
		}
		outProducts = append(outProducts, product)
	}
	return outProducts, nil
}

func (pp *postgresProvider) GetProductsByCategory(ctx context.Context, catCode string) ([]models.Product, error) {
	rows, err := pp.dbConn.Query(ctx, fmt.Sprintf(`
SELECT p.product_id, p.name, p.description, array_agg(c.code) as category_codes
FROM "%s" as p
LEFT JOIN "%s" as pc ON pc.product_id = p.product_id
Left JOIN "%s" as c ON c.category_id = pc.category_id
GROUP BY p.product_id
HAVING $1 = ANY (array_agg(c.code));`,
	pp.cfg.ProductTable,
	pp.cfg.ProductCategoryTable,
	pp.cfg.CatogoryTable),
	catCode)
	if err != nil {
		return nil, ErrQuery
	}
	var outProducts []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.Id, &product.Name, &product.Description, &product.CategoryСodes)
		if err != nil {
			return nil, ErrQuery
		}
		outProducts = append(outProducts, product)
	}
	return outProducts, nil
}

func (pp *postgresProvider) UpdateProductById(ctx context.Context, prodId string, newPorductdata models.ProductForPatch, catIds []int) error {
	transaction, err := pp.dbConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ErrStartTx
	}
	//TODO: rewritre it, hotfix
	if newPorductdata.Name != nil || newPorductdata.Description != nil {
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
		preparedQuery = preparedQuery[:len(preparedQuery)-2]
		usedData = append(usedData, prodId)
		_, err = transaction.Exec(ctx, fmt.Sprintf("%s WHERE \"product_id\" = $%d", preparedQuery, usedFields+1), usedData...)
		if err != nil {
			if err := transaction.Rollback(ctx); err != nil {
				return ErrRollbackTx
			}
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return ErrProductExist
				}
			}
			return ErrQuery
		}
	}
	
	if catIds == nil {
		if err := transaction.Commit(ctx); err != nil {
			return ErrCommitTx
		}
		return nil
	}

	for _, cId := range catIds {
		_, err = transaction.Exec(ctx, fmt.Sprintf(`
INSERT INTO "%s" (product_id, category_id)
VALUES ($1, $2)  ON CONFLICT (product_id, category_id) DO NOTHING;`,
		pp.cfg.ProductCategoryTable),
		prodId,
		cId)
		if err != nil {
			if err := transaction.Rollback(ctx); err != nil {
				return ErrRollbackTx
			}
			return ErrQuery
		}
	}
	
	_, err = transaction.Exec(ctx, fmt.Sprintf(`
DELETE FROM "%s"
WHERE product_id = $1 AND category_id <> ANY ($2);`,
	pp.cfg.ProductCategoryTable),
	prodId,
	catIds)
	if err != nil {
		if err := transaction.Rollback(ctx); err != nil {
			return ErrRollbackTx
		}
		return ErrQuery
	}
	if err := transaction.Commit(ctx); err != nil {
		return ErrCommitTx
	}
	return nil
}

func (pp *postgresProvider) DeleteProductById(ctx context.Context, id string) error {
	transaction, err := pp.dbConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ErrStartTx
	}
	_, err = transaction.Exec(ctx, fmt.Sprintf("DELETE FROM \"%s\" WHERE \"product_id\" = $1", pp.cfg.ProductTable), id)
	if err != nil {
		if err := transaction.Rollback(ctx); err != nil {
			return ErrRollbackTx
		}
		return ErrQuery
	}
	if err := transaction.Commit(ctx); err != nil {
		return ErrCommitTx
	}
	return nil
}
