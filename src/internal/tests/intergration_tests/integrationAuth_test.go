//go:build integration

package integration_tests

import (
	auth_service "annotater/internal/bl/auth"
	user_repo_adapter "annotater/internal/bl/userService/userRepo/userRepoAdapter"
	models_da "annotater/internal/models/modelsDA"
	auth_utils "annotater/internal/pkg/authUtils"
	unit_test_utils "annotater/internal/tests/utils"
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ITAuthTestSuite struct {
	suite.Suite
}

func createDBInContainer(t provider.T) (testcontainers.Container, *gorm.DB) {
	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
			"POSTGRES_DB":       "test_db",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	t.Require().NoError(err)

	// Get the host and port for the database
	host, err := pgContainer.Host(ctx)
	t.Require().NoError(err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	t.Require().NoError(err)

	// Open a new database connection for each test
	dsn := fmt.Sprintf("host=%s port=%s user=test_user password=test_password dbname=test_db sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	t.Require().NoError(err)

	// Automatically migrate the schema for each test, check for errors
	err = db.AutoMigrate(&models_da.User{})
	t.Require().NoError(err) // Ensure migration succeeds

	return pgContainer, db
}

func destroyContainer(t provider.T, pgContainer testcontainers.Container) {
	// Cleanup the container after each test
	ctx := context.Background()
	err := pgContainer.Terminate(ctx)
	t.Require().NoError(err)
}

func (suite *ITAuthTestSuite) TestUsecaseSignUp(t provider.T) {
	// Ensure db is not nil before proceeding
	container, db := createDBInContainer(t)
	defer destroyContainer(t, container)
	t.Require().NotNil(db)

	userRepo := user_repo_adapter.NewUserRepositoryAdapter(db)
	hasher := auth_utils.NewPasswordHashCrypto()
	tokenHandler := auth_utils.NewJWTTokenHandler()
	userService := auth_service.NewAuthService(unit_test_utils.MockLogger, userRepo, hasher, tokenHandler, unit_test_utils.TEST_HASH_KEY)

	userMother := unit_test_utils.NewUserObjectMother()
	user := userMother.DefaultUser()

	err := userService.SignUp(user)
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
	container, db := createDBInContainer(t)
	defer destroyContainer(t, container)
	t.Require().NotNil(db)

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
	var err error
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
