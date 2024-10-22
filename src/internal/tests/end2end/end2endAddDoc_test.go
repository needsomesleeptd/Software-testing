//go:build e2e

package end2end_test

import (
	auth_handler "annotater/internal/http-server/handlers/auth"
	models_dto "annotater/internal/models/dto"
	end2end_utils "annotater/internal/tests/end2end/utils"
	unit_test_utils "annotater/internal/tests/utils"
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type E2ESuite struct {
	suite.Suite
	e           httpexpect.Expect
	done        chan os.Signal
	db          *gorm.DB
	pgContainer testcontainers.Container
	wg          sync.WaitGroup
}

// SetupSuite runs before any tests in the suite
func (suite *E2ESuite) BeforeEach(t provider.T) {
	ctx := context.Background()
	fmt.Print("STARTING SUITE")
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
	t.Require().NoError(err)

	err = end2end_utils.TablesMigrate(db)
	t.Require().NoError(err)

	suite.db = db
	// Initialize httpexpect with a new test.T
	suite.e = *httpexpect.WithConfig(httpexpect.Config{
		Client:   &http.Client{},
		BaseURL:  "http://localhost:8080",
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
	done := make(chan os.Signal, 1)
	suite.done = done
	suite.wg.Add(1)
	go end2end_utils.RunTheApp(suite.db, suite.done, os.Getenv("CONFIG_PATH"), &suite.wg)
}

func (s *E2ESuite) AfterEach(t provider.T) {
	ctx := context.Background()
	err := s.pgContainer.Terminate(ctx)
	t.Require().NoError(err)

	s.done <- os.Interrupt
}

func (s *E2ESuite) Test_E2ELoadingDocument(t provider.T) {
	s.wg.Wait()

	if os.Getenv("INTEGRATION_FAILED") != "" {
		t.XSkip()
	}

	user := unit_test_utils.NewUserObjectMother().DefaultUser()
	userDTO := models_dto.ToDtoUser(*user)
	reqSignUp := auth_handler.RequestSignUp{User: *userDTO}
	fmt.Print("starting test")
	// Assert register response
	s.e.POST("/user/SignUp").
		WithJSON(reqSignUp).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		NotEmpty().
		HasValue("status", "OK")

	reqSignIn := auth_handler.RequestSignIn{Login: user.Login, Password: user.Password}

	s.e.POST("/user/SignIn").
		WithJSON(reqSignIn).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		NotEmpty().
		ContainsKey("Response").
		ContainsKey("jwt")
}

// TestEnd2End runs the test suite
func TestEnd2End(t *testing.T) {
	suite.RunSuite(t, new(E2ESuite))
}
