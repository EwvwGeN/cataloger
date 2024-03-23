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

	"github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/httpmodels"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	v1 "github.com/EwvwGeN/InHouseAd_assignment/internal/http/v1"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/service"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/service/mocks"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/storage"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type catgTestSuite struct {
	suite.Suite
	cfg config.Config
	categoryRepoMock *mocks.CategoryRepo
	addHandler http.HandlerFunc
	editHandler http.HandlerFunc
	deletehHanlder http.HandlerFunc
	getOneHandler http.HandlerFunc
	getAllHandler http.HandlerFunc
}

func TestCategorySuiteRun(t *testing.T) {
	suite.Run(t, new(catgTestSuite))
}

func (suite *catgTestSuite) SetupSuite() {
	cfg := config.Config{
		Validator: config.Validator{
			CategoryNameValidate: `([а-яА-я\w ]+)`,
			CategoryCodeValidate: `^([^\W_]+_?[^\W_])+$`,
			CategoryDescValidate: `([а-яА-я\w ]+)`,
		},
	}
	suite.cfg = cfg
	suite.categoryRepoMock = mocks.NewCategoryRepo(suite.T())
	lg := slog.New(
		slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelError}),
	)
	categoryService := service.NewCategoryService(lg, suite.categoryRepoMock)
	suite.addHandler = v1.CategoryAdd(lg, cfg.Validator, categoryService)
	suite.editHandler = v1.CategoryEdit(lg, cfg.Validator, categoryService)
	suite.deletehHanlder = v1.CategoryDelete(lg, categoryService)
	suite.getOneHandler = v1.CategoryGetOne(lg, categoryService)
	suite.getAllHandler = v1.CategoryGetAll(lg, categoryService)
}

func (suite *catgTestSuite) Test_Add() {
	var categories = map[string]models.Category{}
	tests := []struct{
		name string
		req httpmodels.CategoryAddRequest
		wantSave bool
		wantCode int
	}{
		{
			name: "add_category_happy_pass",
			req: httpmodels.CategoryAddRequest{
				Category: models.Category{
					Name: "Тестовое навзание",
					Code: "test_category",
					Description: "Тествое описание",
				},
			},
			wantSave: true,
			wantCode: http.StatusCreated,
		},
		{
			name: "add_category_happy_pass",
			req: httpmodels.CategoryAddRequest{
				Category: models.Category{
					Name: "Тестовое навзание 2",
					Code: "test_category",
					Description: "Тествое описание 2",
				},
			},
			wantSave: true,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "add_category_with_wrong_name",
			req: httpmodels.CategoryAddRequest{
				Category: models.Category{
					Name: "",
				},
			},
			wantSave: false,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "add_category_with_wrong_code",
			req: httpmodels.CategoryAddRequest{
				Category: models.Category{
					Name: "Cool test name",
					Code: "",
				},
			},
			wantSave: false,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "add_category_with_wrong_description",
			req: httpmodels.CategoryAddRequest{
				Category: models.Category{
					Name: "Cool test name",
					Code: "cool_test_code",
					Description: "",
				},
			},
			wantSave: false,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "add_category_with_not_valid_code",
			req: httpmodels.CategoryAddRequest{
				Category: models.Category{
					Name: "Cool test name",
					Code: "test code",
				},
			},
			wantSave: false,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		err := json.NewEncoder(&jsonBody).Encode(&tt.req)
		suite.Require().NoError(err, "failed to encode request")
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/category/add", &jsonBody)
		if tt.wantSave {
			suite.categoryRepoMock.On("SaveCategory", mock.Anything, tt.req.Category).Once().
			Return(func(ctx context.Context, category models.Category) error {
				if _, ok := categories[category.Code]; ok{
					return storage.ErrCategoryExist
				}
				categories[category.Code] = category
				return nil
			})
		}
		suite.addHandler.ServeHTTP(w,r)
		suite.Require().Equal(tt.wantCode, w.Code, "test: %s", tt.name)
	}
}

func (suite *catgTestSuite) Test_Edit() {
	categories := map[string]models.Category{
		"test_category_one": {
			Name: "First test category",
			Code: "test_category_one",
			Description: "First description",
		},
	}
	tests := []struct{
		name string
		catCode string
		req httpmodels.CategoryEditRequest
		wantEdit bool
		wantCode int
	}{
		{
			name: "happy_edit",
			catCode: "test_category_one",
			req: httpmodels.CategoryEditRequest{
				CategoryNewData: models.CategoryForPatch{
					Name: func () *string {
						name := "New name"
						return &name
					}(),
					Code: func () *string {
						code := "new_code"
						return &code
					}(),
					Description: func () *string {
						description := "New description"
						return &description
					}(),
				},
			},
			wantEdit: true,
			wantCode: http.StatusOK,
		},
		{
			name: "happy_edit_without_any_fields",
			catCode: "test_category_one",
			req: httpmodels.CategoryEditRequest{
				CategoryNewData: models.CategoryForPatch{
					Name: nil,
					Code: func () *string {
						code := "new_code"
						return &code
					}(),
					Description: func () *string {
						description := "New description"
						return &description
					}(),
				},
			},
			wantEdit: true,
			wantCode: http.StatusOK,
		},
		{
			name: "edit_without_fields",
			catCode: "test_category_one",
			req: httpmodels.CategoryEditRequest{
				CategoryNewData: models.CategoryForPatch{
					Name: nil,
					Code: nil,
					Description: nil,
				},
			},
			wantEdit: false,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "not_valid_new_data",
			catCode: "test_category_one",
			req: httpmodels.CategoryEditRequest{
				CategoryNewData: models.CategoryForPatch{
					Name: nil,
					Code: func () *string {
						code := "new code"
						return &code
					}(),
					Description: nil,
				},
			},
			wantEdit: false,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests{
		var jsonBody bytes.Buffer
		err := json.NewEncoder(&jsonBody).Encode(&tt.req)
		suite.Require().NoError(err, "failed to encode request")
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/category/%s/edit", tt.catCode)
		r := httptest.NewRequest(http.MethodPost, url, &jsonBody)
		vars := map[string]string{
			"catCode": tt.catCode,
		}
		r = mux.SetURLVars(r, vars)
		if tt.wantEdit {
			suite.categoryRepoMock.On("UpdateCategoryByCode", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().
			Return(func(ctx context.Context, catCode string, updateData models.CategoryForPatch) error {
				_, ok := categories[catCode]
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

func (suite *catgTestSuite) Test_Delete() {
	tests := []struct{
		name string
		catCode string
		wantCode int
		haveCatCode bool
		wantDelete bool
	}{
		{
			name: "happy_pass",
			catCode: "test_code",
			wantDelete: true,
			haveCatCode: true,
			wantCode: http.StatusOK,
		},
		{
			name: "without_cat_code",
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/category/%s/delete", tt.catCode)
		r := httptest.NewRequest(http.MethodDelete, url, &jsonBody)
		var vars map[string]string
		if tt.haveCatCode {
			vars = map[string]string{
				"catCode": tt.catCode,
			}
		}
		r = mux.SetURLVars(r, vars)
		if tt.wantDelete {
			suite.categoryRepoMock.On("DeleteCategoryBycode", mock.Anything, mock.AnythingOfType("string")).
			Once().Return(nil)
		}
		suite.deletehHanlder.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code, "test: %s", tt.name)
	}
}

func (suite *catgTestSuite) Test_GetOne() {
	categories := map[string]models.Category{
		"test_category_one": {
			Name: "Test category",
			Code: "test_category_one",
			Description: "Test category for GetOne",
		},
	}
	tests := []struct{
		name string
		catCode string
		haveCatCode bool
		wantGet bool
		wantCode int
	}{
		{
			name: "happy_pass",
			catCode: "test_category_one",
			haveCatCode: true,
			wantGet: true,
			wantCode: http.StatusOK,
		},
		{
			name: "without_code",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "wrong_code",
			catCode: "test_wrong_category_code",
			haveCatCode: true,
			wantGet: true,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/category/%s", tt.catCode)
		r := httptest.NewRequest(http.MethodGet, url, &jsonBody)
		var vars map[string]string
		if tt.haveCatCode {
			vars = map[string]string{
				"catCode": tt.catCode,
			}
		}
		r = mux.SetURLVars(r, vars)
		if tt.wantGet {
			suite.categoryRepoMock.On("GetCategoryByCode", mock.Anything, mock.AnythingOfType("string")).
			Once().Return(func(cxt context.Context, catCode string) (models.Category, error) {
				catg, ok := categories[catCode]
				if !ok {
					return models.Category{}, storage.ErrQuery
				}
				return catg, nil
			})
		}
		suite.getOneHandler.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code, "test: %s", tt.name)
		if tt.wantCode == http.StatusOK {
			var resp httpmodels.CategoryGetOneResponse
			err := json.NewDecoder(w.Body).Decode(&resp)
			suite.Require().NoError(err, "test name: %s", tt.name)
			suite.Require().Equal(categories[tt.catCode], resp.Category)
		}
	}
}

func (suite *catgTestSuite) Test_GetAll() {
	categories := map[string]models.Category{
		"test_category_one": {
			Name: "Test category",
			Code: "test_category_one",
			Description: "Test category for GetAll",
		},
		"test_category_two": {
			Name: "Test category",
			Code: "test_category_two",
			Description: "Test category for GetAll",
		},
	}
	tests := []struct{
		name string
		catCode string
		haveCatCode bool
		wantGet bool
		wantCode int
	}{
		{
			name: "happy_pass",
			catCode: "test_category_one",
			haveCatCode: true,
			wantGet: true,
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/categories", &jsonBody)
		if tt.wantGet {
			suite.categoryRepoMock.On("GetAllCategories", mock.Anything).
			Once().Return(func(cxt context.Context) ([]models.Category, error) {
				var outCatgs []models.Category
				for _, catg := range categories {
					outCatgs = append(outCatgs, catg)
				}
				return outCatgs, nil
			})
		}
		suite.getAllHandler.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code, "test: %s", tt.name)
		if tt.wantCode == http.StatusOK {
			var resp httpmodels.CategoryGetAllResponse
			err := json.NewDecoder(w.Body).Decode(&resp)
			suite.Require().NoError(err, "test name: %s", tt.name)
			suite.Require().Equal(func() ([]models.Category) {
				var outCatgs []models.Category
				for _, catg := range categories {
					outCatgs = append(outCatgs, catg)
				}
				return outCatgs
			}(), resp.Categories)
		}
	}
}