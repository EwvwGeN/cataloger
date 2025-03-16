package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EwvwGeN/cataloger/internal/config"
	"github.com/EwvwGeN/cataloger/internal/domain/httpmodels"
	"github.com/EwvwGeN/cataloger/internal/domain/models"
	v1 "github.com/EwvwGeN/cataloger/internal/http/v1"
	"github.com/EwvwGeN/cataloger/internal/service"
	"github.com/EwvwGeN/cataloger/internal/service/mocks"
	"github.com/EwvwGeN/cataloger/internal/storage"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type prodTestSuite struct {
	suite.Suite
	cfg              config.Config
	categoryCodesRepoMock *mocks.CategoryCodesRepo
	productRepoMock *mocks.ProductRepo
	addHandler       http.HandlerFunc
	editHandler      http.HandlerFunc
	deletehHanlder   http.HandlerFunc
	getOneHandler    http.HandlerFunc
	getAllHandler    http.HandlerFunc
	getAllByCategoryHandler    http.HandlerFunc
}

func TestProductSuiteRun(t *testing.T) {
	suite.Run(t, new(prodTestSuite))
}

func (suite *prodTestSuite) SetupSuite() {
	suite.cfg = config.Config{
		Validator: config.Validator{
			ProductNameValidate: `([а-яА-я\w ]+)`,
			ProductDescValidate: `([а-яА-я\w ]+)`,
		},
	}
	suite.categoryCodesRepoMock = mocks.NewCategoryCodesRepo(suite.T())
	suite.productRepoMock = mocks.NewProductRepo(suite.T())
	lg := slog.New(
		slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelError}),
	)
	productServive := service.NewProductService(lg, suite.productRepoMock, suite.categoryCodesRepoMock)
	suite.addHandler = v1.ProductAdd(lg, suite.cfg.Validator, productServive)
	suite.editHandler = v1.ProductEdit(lg, suite.cfg.Validator, productServive)
	suite.deletehHanlder = v1.ProductDelete(lg, productServive)
	suite.getOneHandler = v1.ProductGetOne(lg, productServive)
	suite.getAllHandler = v1.ProductGetAll(lg, productServive)
	suite.getAllByCategoryHandler = v1.ProductGetAllByCategory(lg, productServive)
}

func (suite *prodTestSuite) Test_Add() {
	categories := map[string]int{
		"test_category_one": 1,
		"test_category_two": 2,
		"test_category_three": 3,
	}
	products := map[string]struct{}{
		"Second product": {},
	}
	tests := []struct{
		name string
		newProdId string
		req httpmodels.ProductAddRequest
		wantGetCategoryId bool
		wantSave bool
		wantCode int
	}{
		{
			name: "happy_pass",
			newProdId: "1",
			req: httpmodels.ProductAddRequest{
				Product: models.Product{
					Name: "New product",
					Description: "Description for product",
					CategoryСodes: []string{
						"test_category_one",
						"test_category_two",
					},
				},
			},
			wantGetCategoryId: true,
			wantSave: true,
			wantCode: http.StatusCreated,
		},
		{
			name: "product_already_exist",
			req: httpmodels.ProductAddRequest{
				Product: models.Product{
					Name: "Second product",
					Description: "Description for product",
					CategoryСodes: []string{
						"test_category_one",
						"test_category_two",
					},
				},
			},
			wantGetCategoryId: true,
			wantSave: true,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "not_valid_name",
			req: httpmodels.ProductAddRequest{
				Product: models.Product{
					Name: "---",
					Description: "Description for product",
					CategoryСodes: []string{
						"test_category_one",
						"test_category_two",
					},
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "not_valid_description",
			req: httpmodels.ProductAddRequest{
				Product: models.Product{
					Name: "New product",
					Description: "---",
					CategoryСodes: []string{
						"test_category_one",
						"test_category_two",
					},
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "product_category_not_exist",
			req: httpmodels.ProductAddRequest{
				Product: models.Product{
					Name: "Second product",
					Description: "Description for product",
					CategoryСodes: []string{
						"test_category_one",
						"test_category_two",
						"test_category_five",
					},
				},
			},
			wantGetCategoryId: true,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		var jsonBody bytes.Buffer
		err := json.NewEncoder(&jsonBody).Encode(&tt.req)
		suite.Require().NoError(err, "failed to encode request")
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/product/add", &jsonBody)
		if tt.wantGetCategoryId {
			suite.categoryCodesRepoMock.On("GetCategoriesIdByCodes", mock.Anything, tt.req.Product.CategoryСodes).Once().
			Return(func(ctx context.Context, catCodes []string) ([]int, error) {
				var categoriesId []int
				for _, code := range catCodes {
					if catId, ok := categories[code]; ok {
						categoriesId = append(categoriesId, catId)
					}
				}
				return categoriesId, nil
			})
		}
		if tt.wantSave {
			suite.productRepoMock.On("SaveProduct", mock.Anything, mock.Anything, mock.Anything).Once().
			Return(func(ctx context.Context, product models.Product, catIds []int) (string, error) {
				if _, ok := products[product.Name]; ok {
					return "", storage.ErrProductExist
				}
				return tt.newProdId, nil
			})
		}
		suite.addHandler.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code)
		if tt.wantCode == http.StatusCreated {
			var resp httpmodels.ProductAddResponse
			err := json.NewDecoder(w.Body).Decode(&resp)
			suite.Require().NoError(err)
			suite.Require().Equal(tt.newProdId, resp.ProductId)
		}
	}
}

func (suite *prodTestSuite) Test_Edit() {
	products := map[string]struct{}{
		"1": {},
	}
	categories := map[string]int{
		"test_category_one": 1,
		"test_category_two": 2,
	}
	tests := []struct{
		name string
		prodId string
		req httpmodels.ProductEditRequest
		wantGetCategoryId bool
		wantEdit bool
		wantCode int
	}{
		{
			name: "happy_pass",
			prodId: "1",
			req: httpmodels.ProductEditRequest{
				ProductNewData: models.ProductForPatch{
					Name: func () *string {
						name := "New name"
						return &name
					}(),
					Description: func () *string {
						desc := "New description"
						return &desc
					}(),
					CategoryСodes: []string{
						"test_category_one",
					},
				},
			},
			wantGetCategoryId: true,
			wantEdit: true,
			wantCode: http.StatusOK,
		},
		{
			name: "not_valid_categories",
			prodId: "1",
			req: httpmodels.ProductEditRequest{
				ProductNewData: models.ProductForPatch{
					Name: func () *string {
						name := "New name"
						return &name
					}(),
					Description: func () *string {
						desc := "New description"
						return &desc
					}(),
					CategoryСodes: []string{
						"test_category_five",
					},
				},
			},
			wantGetCategoryId: true,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "not_valid_name",
			prodId: "1",
			req: httpmodels.ProductEditRequest{
				ProductNewData: models.ProductForPatch{
					Name: func () *string {
						name := "---"
						return &name
					}(),
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "not_valid_description",
			prodId: "1",
			req: httpmodels.ProductEditRequest{
				ProductNewData: models.ProductForPatch{
					Name: func () *string {
						name := "AAA"
						return &name
					}(),
					Description: func () *string {
						desc := "---"
						return &desc
					}(),
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "nothing_to_update",
			prodId: "1",
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		err := json.NewEncoder(&jsonBody).Encode(&tt.req)
		suite.Require().NoError(err, "failed to encode request")
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/product/%s/edit", tt.prodId)
		r := httptest.NewRequest(http.MethodPatch, url, &jsonBody)
		vars := map[string]string{
			"productId": tt.prodId,
		}
		r = mux.SetURLVars(r, vars)
		if tt.wantGetCategoryId {
			suite.categoryCodesRepoMock.On("GetCategoriesIdByCodes", mock.Anything, tt.req.ProductNewData.CategoryСodes).Once().
			Return(func(ctx context.Context, catCodes []string) ([]int, error) {
				var categoriesId []int
				for _, code := range catCodes {
					if catId, ok := categories[code]; ok {
						categoriesId = append(categoriesId, catId)
					}
				}
				return categoriesId, nil
			})
		}
		if tt.wantEdit {
			suite.productRepoMock.On("UpdateProductById", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
			Return(func(ctx context.Context, prodID string, updateData models.ProductForPatch, catIds []int) error {
				_, ok := products[prodID]
				if !ok {
					return storage.ErrQuery
				}
				return nil
			})
		}
		suite.editHandler.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code, "test: %s", tt.name)
	}
}

func (suite *prodTestSuite) Test_GetOne() {
	products := map[string]models.Product{
		"1": {
			Name: "Test product",
			Description: "test product",
			CategoryСodes: []string{
				"test_category_one",
			},
		},
	}
	tests := []struct{
		name string
		prodId string
		wantGet bool
		wantCode int
	}{
		{
			name: "happy_pass",
			prodId: "1",
			wantGet: true,
			wantCode: http.StatusOK,
		},
		{
			name: "product_not_exist",
			prodId: "2",
			wantGet: true,
			wantCode: http.StatusOK,
		},
		{
			name: "without_product_id",
			prodId: "",
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/product/%s", tt.prodId)
		r := httptest.NewRequest(http.MethodGet, url, &jsonBody)
		vars := map[string]string{
			"productId": tt.prodId,
		}
		r = mux.SetURLVars(r, vars)
		if tt.wantGet {
			suite.productRepoMock.On("GetProductById", mock.Anything, mock.Anything).Once().
			Return(func(ctx context.Context, prodId string) (models.Product, error) {
				return products[prodId], nil
			})
		}
		suite.getOneHandler.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code)
		if tt.wantCode == http.StatusOK {
			var resp httpmodels.ProductGetOneResponse
			err := json.NewDecoder(w.Body).Decode(&resp)
			suite.Require().NoError(err)
			suite.Require().Equal(products[tt.prodId], resp.Product)
		}
	}
}

func (suite *prodTestSuite) Test_GetAll() {
	products := map[string]models.Product{
		"1": {
			Name: "Test product",
			Description: "test product",
			CategoryСodes: []string{
				"test_category_one",
			},
		},
		"2": {
			Name: "Test product",
			Description: "test product",
			CategoryСodes: []string{
				"test_category_one",
			},
		},
	}
	tests := []struct{
		name string
		wantGet bool
		wantCode int
	}{
		{
			name: "happy_pass",
			wantGet: true,
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/products", &jsonBody)
		if tt.wantGet {
			suite.productRepoMock.On("GetAllProducts", mock.Anything).Once().
			Return(func(ctx context.Context) ([]models.Product, error) {
				var outProducts []models.Product
				for _, p := range products {
					outProducts = append(outProducts, p)
				}
				return outProducts, nil
			})
		}
		suite.getAllHandler.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code)
		if tt.wantCode == http.StatusOK {
			var resp httpmodels.ProductGetAllResponse
			err := json.NewDecoder(w.Body).Decode(&resp)
			suite.Require().NoError(err)
			var checkProducts []models.Product
			for _, p := range products {
				checkProducts = append(checkProducts, p)
			}
			suite.Require().Equal(checkProducts, resp.Products)
		}
	}
}

func (suite *prodTestSuite) Test_GetAllByCategory() {
	products := map[string][]models.Product{
		"test_category_one": {
			{
				Name: "Test product",
				Description: "test product",
				CategoryСodes: []string{
					"test_category_one",
				},
			},
			{
				Name: "Test product 2",
				Description: "test product",
				CategoryСodes: []string{
					"test_category_one",
				},
			},
		},
	}
	tests := []struct{
		name string
		categoryCode string
		wantGet bool
		wantCode int
	}{
		{
			name: "happy_pass",
			categoryCode: "test_category_one",
			wantGet: true,
			wantCode: http.StatusOK,
		},
		{
			name: "without_category_code",
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/products/%s", tt.categoryCode)
		r := httptest.NewRequest(http.MethodPatch, url, &jsonBody)
		vars := map[string]string{
			"catCode": tt.categoryCode,
		}
		r = mux.SetURLVars(r, vars)
		if tt.wantGet {
			suite.productRepoMock.On("GetProductsByCategory", mock.Anything, mock.Anything).Once().
			Return(func(ctx context.Context, catCode string) ([]models.Product, error) {
				return products[catCode], nil
			})
		}
		suite.getAllByCategoryHandler.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code)
		if tt.wantCode == http.StatusOK {
			var resp httpmodels.ProductGetAllResponse
			err := json.NewDecoder(w.Body).Decode(&resp)
			suite.Require().NoError(err)
			suite.Require().Equal(products[tt.categoryCode], resp.Products)
		}
	}
}

func (suite *prodTestSuite) Test_Delete() {
	tests := []struct{
		name string
		prodId string
		wantDelete bool
		wantCode int
		
	}{
		{
			name: "happy_pass",
			prodId: "test_code",
			wantDelete: true,
			wantCode: http.StatusOK,
		},
		{
			name: "without_prod_code",
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/product/%s/delete", tt.prodId)
		r := httptest.NewRequest(http.MethodDelete, url, &jsonBody)
		vars := map[string]string{
			"productId": tt.prodId,
		}
		r = mux.SetURLVars(r, vars)
		if tt.wantDelete {
			suite.productRepoMock.On("DeleteProductById", mock.Anything, mock.AnythingOfType("string")).
			Once().Return(nil)
		}
		suite.deletehHanlder.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code, "test: %s", tt.name)
	}
}