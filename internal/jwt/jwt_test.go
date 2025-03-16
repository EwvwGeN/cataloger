package jwt

import (
	"testing"
	"time"

	"github.com/EwvwGeN/cataloger/internal/domain/models"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	suite.Suite
	secretKey string
	jwtManager *jwtManager
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (suite *testSuite) SetupSuite() {
	suite.secretKey = "test_secret_key"
	suite.jwtManager = NewJwtManager(suite.secretKey)
}

func (suite *testSuite) Test_CreateTokenHappyPass(){
	user := models.User{
		Email: "test@test.test",
	}
	duration := time.Duration(10*time.Second)
	token, err := suite.jwtManager.CreateJWT(user, duration)
	creatingTime := time.Now()
	suite.Require().NoError(err)
	suite.Require().NotEmpty(token)
	claims, err := suite.jwtManager.MustParseJwt(token)
	suite.Require().NoError(err)
	suite.Equal(user.Email, claims["email"].(string))
	suite.InDelta(creatingTime.Add(duration).Unix(), claims["exp"].(float64), 1)
}

func (suite *testSuite) Test_CreateTokenEmptyEmail(){
	user := models.User{
	}
	duration := time.Duration(10*time.Second)
	token, err := suite.jwtManager.CreateJWT(user, duration)
	suite.Require().Error(err)
	suite.Require().Empty(token)
}

func (suite *testSuite) Test_CreateRefreshHappyPass(){
	refresh, err := suite.jwtManager.CreateRefresh()
	suite.Require().NoError(err)
	suite.Require().NotEmpty(refresh)
}