package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/EwvwGeN/cataloger/internal/domain/models"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

func (pp *postgresProvider) SaveCategory(ctx context.Context, category models.Category) error {
	_, err := pp.dbConn.Exec(ctx, fmt.Sprintf(`INSERT INTO "%s" (name, code, description)
VALUES($1,$2,$3);`,
	pp.cfg.CatogoryTable),
	category.Name,
	category.Code,
	category.Description)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrCategoryExist
		}
	}
	return ErrQuery
}

func (pp *postgresProvider) InserOrGetCategiriesId(ctx context.Context, categories []models.Category) (map[string]int, error) {
	transaction, err := pp.dbConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, ErrStartTx
	}
	categoriesMap := make(map[string]int, len(categories))
	for _, catg := range categories {
		var id int
		err = transaction.QueryRow(ctx, fmt.Sprintf(`
WITH ins AS(
	INSERT INTO "%s" (name, code, description)
			VALUES ($1, $2, $3) 
	ON CONFLICT ("code") DO NOTHING
	RETURNING category_id
)
SELECT category_id FROM ins
UNION
	SELECT category_id FROM "%s" WHERE code = $2;`,
		pp.cfg.CatogoryTable, pp.cfg.CatogoryTable),
		catg.Name,
		catg.Code,
		catg.Description).
		Scan(&id)
		if err != nil {
			if err := transaction.Rollback(ctx); err != nil {
				return nil, ErrRollbackTx
			}
			return nil, ErrQuery
		}
		categoriesMap[catg.Code] = id
	}
	if err := transaction.Commit(ctx); err != nil {
		return nil, ErrCommitTx
	}
	return categoriesMap, nil
}

func (pp *postgresProvider) GetCategoryByCode(ctx context.Context, catCode string) (models.Category, error) {
	row := pp.dbConn.QueryRow(ctx, fmt.Sprintf(`
SELECT "name", "code", "description"
FROM "%s"
WHERE "code"=$1;`,
	pp.cfg.CatogoryTable),
	catCode)
	var (
		category models.Category
	)
	err := row.Scan(&category.Name, &category.Code, &category.Description)
	if err != nil {
		return models.Category{}, ErrQuery
	}
	return category, nil
}

func (pp *postgresProvider) GetCategoriesIdByCodes(ctx context.Context, catCodes []string) ([]int, error) {
	row:= pp.dbConn.QueryRow(ctx, fmt.Sprintf(`
SELECT array_agg(category_id)
FROM "%s"
WHERE code = ANY ($1)`,
	pp.cfg.CatogoryTable),
	catCodes)
	var outCategoriesId []int
	err := row.Scan(&outCategoriesId)
	if err != nil {
		return nil, ErrQuery
	}
	return outCategoriesId, nil
}

func (pp *postgresProvider) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := pp.dbConn.Query(ctx, fmt.Sprintf(`
SELECT "name", "code", "description"
FROM "%s"`,
	pp.cfg.CatogoryTable))
	if err != nil {
		return nil, ErrQuery
	}
	var outCategorys []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.Name, &category.Code, &category.Description)
		if err != nil {
			return nil, ErrQuery
		}
		outCategorys = append(outCategorys, category)
	}
	return outCategorys, nil
}

func (pp *postgresProvider) UpdateCategoryByCode(ctx context.Context, catCode string, catUpdateData models.CategoryForPatch) error {
	preparedQuery := fmt.Sprintf("UPDATE \"%s\" SET ", pp.cfg.CatogoryTable)
	// is it faster to use marshal to json and unmarshal to map[string]interface{} and then range it by for statement?
	usedFields := 0
	usedData := make([]interface{}, 0)
	if catUpdateData.Name != nil {
		preparedQuery += fmt.Sprintf("\"name\" = $%d, ", usedFields+1)
		usedFields++
		usedData = append(usedData, *catUpdateData.Name)
	}
	if catUpdateData.Code != nil {
		preparedQuery += fmt.Sprintf("\"code\" = $%d, ", usedFields+1)
		usedFields++
		usedData = append(usedData, *catUpdateData.Code)
	}
	if catUpdateData.Description != nil {
		preparedQuery += fmt.Sprintf("\"description\" = $%d, ", usedFields+1)
		usedFields++
		usedData = append(usedData, *catUpdateData.Description)
	}
	// the worst but fast solution
	preparedQuery = preparedQuery[:len(preparedQuery)-2]
	usedData = append(usedData, catCode)
	_, err := pp.dbConn.Exec(ctx, fmt.Sprintf("%s WHERE \"code\" = $%d", preparedQuery, usedFields+1), usedData...)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrCategoryExist
		}
	}
	return ErrQuery
}

func (pp *postgresProvider) DeleteCategoryBycode(ctx context.Context, catCode string) error {
	_, err := pp.dbConn.Exec(ctx, fmt.Sprintf("DELETE FROM \"%s\" WHERE \"code\" = $1", pp.cfg.CatogoryTable), catCode)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23503" {
			return ErrCategoryUsed
		}
	}
	return nil
}
