//go:build inegration

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

	"github.com/stretchr/testify/suite"
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

func (suite *ITAuthTestSuite) SetupTest() {
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
	suite.Require().NoError(err)
	suite.pgContainer = pgContainer

	// Get the host and port for the database
	host, err := pgContainer.Host(ctx)
	suite.Require().NoError(err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	suite.Require().NoError(err)

	// Open a new database connection for each test
	dsn := fmt.Sprintf("host=%s port=%s user=test_user password=test_password dbname=test_db sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	suite.Require().NoError(err)

	// Automatically migrate the schema for each test
	db.AutoMigrate(&models_da.User{}) // fix migrations
	suite.db = db
}

func (suite *ITAuthTestSuite) TearDownTest() {
	// Cleanup the container after each test
	ctx := context.Background()
	err := suite.pgContainer.Terminate(ctx)
	suite.Require().NoError(err)
}

func (suite *ITAuthTestSuite) TestUsecaseSignUp() {
	userRepo := user_repo_adapter.NewUserRepositoryAdapter(suite.db)
	hasher := auth_utils.NewPasswordHashCrypto()
	tokenHandler := auth_utils.NewJWTTokenHandler()
	userService := auth_service.NewAuthService(unit_test_utils.MockLogger, userRepo, hasher, tokenHandler, unit_test_utils.TEST_HASH_KEY)
	var id uint64 = 1

	userMother := unit_test_utils.NewUserObjectMother()
	user := userMother.DefaultUser()

	err := userService.SignUp(user)
	suite.Require().NoError(err)

	gotUser, err := userRepo.GetUserByLogin(user.Login)
	suite.Require().NoError(err)
	//fmt.Print(user, gotUser)
	suite.Require().NoError(hasher.ComparePasswordhash(user.Password, gotUser.Password))

	var gotUserDa *models_da.User
	suite.Require().NoError(suite.db.Model(&models_da.User{}).Where("id = ?", id).Take(&gotUserDa).Error)
	suite.Assert().Equal(*gotUser, models_da.FromDaUser(gotUserDa))
}

func (suite *ITAuthTestSuite) TestUsecaseSignIn() {
	userRepo := user_repo_adapter.NewUserRepositoryAdapter(suite.db)
	hasher := auth_utils.NewPasswordHashCrypto()
	tokenHandler := auth_utils.NewJWTTokenHandler()

	userService := auth_service.NewAuthService(unit_test_utils.MockLogger, userRepo, hasher, tokenHandler, unit_test_utils.TEST_HASH_KEY)

	userMother := unit_test_utils.NewUserObjectMother()
	user := userMother.DefaultUser()

	err := userService.SignUp(user)
	suite.Require().NoError(err)
	token, err := userService.SignIn(user)
	suite.Require().NoError(err)

	err = tokenHandler.ValidateToken(token, unit_test_utils.TEST_HASH_KEY)
	suite.Require().NoError(err)
}

func TestSuiteAuth(t *testing.T) {
	suite.Run(t, new(ITAuthTestSuite))
}
