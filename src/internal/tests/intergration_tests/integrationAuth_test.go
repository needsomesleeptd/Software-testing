//go:build integration

package integration_tests

import (
	auth_service "annotater/internal/bl/auth"
	user_repo_adapter "annotater/internal/bl/userService/userRepo/userRepoAdapter"
	models_da "annotater/internal/models/modelsDA"
	auth_utils "annotater/internal/pkg/authUtils"
	integration_utils "annotater/internal/tests/intergration_tests/utils"
	unit_test_utils "annotater/internal/tests/utils"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type ITAuthTestSuite struct {
	suite.Suite
}

func (suite *ITAuthTestSuite) TestUsecaseSignUp(t provider.T) {
	if os.Getenv("UNIT_FAILED") != "" {
		t.XSkip()
	}
	// Ensure db is not nil before proceeding
	container, db := integration_utils.CreateDBInContainer(t)
	defer integration_utils.DestroyContainer(t, container)
	t.Require().NotNil(db)
	err := db.AutoMigrate(&models_da.User{})
	t.Require().NoError(err)

	userRepo := user_repo_adapter.NewUserRepositoryAdapter(db)
	hasher := auth_utils.NewPasswordHashCrypto()
	tokenHandler := auth_utils.NewJWTTokenHandler()
	userService := auth_service.NewAuthService(unit_test_utils.MockLogger, userRepo, hasher, tokenHandler, unit_test_utils.TEST_HASH_KEY)

	userMother := unit_test_utils.NewUserObjectMother()
	user := userMother.DefaultUser()

	err = userService.SignUp(user)
	t.Require().NoError(err)

	var gotUserDa *models_da.User
	t.Require().NoError(db.Model(&models_da.User{}).Where("login = ?", user.Login).Take(&gotUserDa).Error)
	t.Require().NoError(hasher.ComparePasswordhash(user.Password, gotUserDa.Password))

	gotUser := models_da.FromDaUser(gotUserDa)
	gotUser.Password = user.Password // i store only hashes in db
	user.ID = gotUser.ID             // have no ID before insertion
	t.Assert().Equal(*user, gotUser)

}

func (suite *ITAuthTestSuite) TestUsecaseSignIn(t provider.T) {
	if os.Getenv("UNIT_FAILED") != "" {
		t.XSkip()
	}

	container, db := integration_utils.CreateDBInContainer(t)
	defer integration_utils.DestroyContainer(t, container)
	t.Require().NotNil(db)
	err := db.AutoMigrate(&models_da.User{})
	t.Require().NoError(err)

	// Setup mock controller
	ctr := gomock.NewController(t)
	defer ctr.Finish() // Ensure that the controller is finished at the end of the test

	// Arrange
	userRepo := user_repo_adapter.NewUserRepositoryAdapter(db)
	hasher := auth_utils.NewPasswordHashCrypto()
	tokenHandler := auth_utils.NewJWTTokenHandler()
	userService := auth_service.NewAuthService(unit_test_utils.MockLogger, userRepo, hasher, tokenHandler, unit_test_utils.TEST_HASH_KEY)

	userMother := unit_test_utils.NewUserObjectMother()
	userSignUp := userMother.DefaultUser()

	userSignIn := userMother.DefaultUser()
	// Not deterministic passwd gen
	//hasher.EXPECT().ComparePasswordhash(userSignIn.Password, unit_test_utils.TEST_HASH_PASSWD).Return(nil)

	// Sign up the user first

	userSignUp.Password, err = hasher.GenerateHash(userSignIn.Password)
	t.Require().NoError(err)

	// Add the signed-up user to the database directly using GORM
	t.Require().NoError(db.Create(&userSignUp).Error)

	// Now attempt to sign in with the user
	token, err := userService.SignIn(userSignIn)
	t.Require().NoError(err)

	// Validate the token
	err = tokenHandler.ValidateToken(token, unit_test_utils.TEST_HASH_KEY)
	t.Require().NoError(err)
}

func TestSuiteAuth(t *testing.T) {
	suite.RunSuite(t, new(ITAuthTestSuite))
}
