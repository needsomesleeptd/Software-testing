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

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ITAuthTestSuite struct {
	suite.Suite
	db          *gorm.DB
	pgContainer testcontainers.Container
}

func (suite *ITAuthTestSuite) BeforeEach(t provider.T) {
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
	suite.pgContainer = pgContainer

	// Get the host and port for the database
	host, err := pgContainer.Host(ctx)
	t.Require().NoError(err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	t.Require().NoError(err)

	// Open a new database connection for each test
	dsn := fmt.Sprintf("host=%s port=%s user=test_user password=test_password dbname=test_db sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	fmt.Printf("got db %v,err %v", db, err)
	t.Require().NoError(err)

	// Automatically migrate the schema for each test, check for errors
	err = db.AutoMigrate(&models_da.User{})
	t.Require().NoError(err) // Ensure migration succeeds

	suite.db = db // Now `suite.db` is initialized successfully
}

func (suite *ITAuthTestSuite) AfterEach(t provider.T) {
	// Cleanup the container after each test
	ctx := context.Background()
	err := suite.pgContainer.Terminate(ctx)
	t.Require().NoError(err)
}

func (suite *ITAuthTestSuite) TestUsecaseSignUp(t provider.T) {
	// Ensure suite.db is not nil before proceeding
	t.Require().NotNil(suite.db)

	userRepo := user_repo_adapter.NewUserRepositoryAdapter(suite.db)
	hasher := auth_utils.NewPasswordHashCrypto()
	tokenHandler := auth_utils.NewJWTTokenHandler()
	userService := auth_service.NewAuthService(unit_test_utils.MockLogger, userRepo, hasher, tokenHandler, unit_test_utils.TEST_HASH_KEY)

	userMother := unit_test_utils.NewUserObjectMother()
	user := userMother.DefaultUser()

	err := userService.SignUp(user)
	t.Require().NoError(err)

	gotUser, err := userRepo.GetUserByLogin(user.Login)
	t.Require().NoError(err)
	t.Require().NoError(hasher.ComparePasswordhash(user.Password, gotUser.Password))

	var gotUserDa *models_da.User
	var id uint64 = gotUser.ID
	t.Require().NoError(suite.db.Model(&models_da.User{}).Where("id = ?", id).Take(&gotUserDa).Error)
	t.Assert().Equal(*gotUser, models_da.FromDaUser(gotUserDa))
}

func (suite *ITAuthTestSuite) TestUsecaseSignIn(t provider.T) {

	t.Require().NotNil(suite.db)

	userRepo := user_repo_adapter.NewUserRepositoryAdapter(suite.db)
	hasher := auth_utils.NewPasswordHashCrypto()
	tokenHandler := auth_utils.NewJWTTokenHandler()
	userService := auth_service.NewAuthService(unit_test_utils.MockLogger, userRepo, hasher, tokenHandler, unit_test_utils.TEST_HASH_KEY)

	userMother := unit_test_utils.NewUserObjectMother()
	user := userMother.DefaultUser()

	err := userService.SignUp(user)
	t.Require().NoError(err)

	token, err := userService.SignIn(user)
	t.Require().NoError(err)

	err = tokenHandler.ValidateToken(token, unit_test_utils.TEST_HASH_KEY)
	t.Require().NoError(err)
}

func TestSuiteAuth(t *testing.T) {
	suite.RunSuite(t, new(ITAuthTestSuite))
}
