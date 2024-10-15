//go:build e2e

package end2end_test

import (
	auth_handler "annotater/internal/http-server/handlers/auth"
	models_dto "annotater/internal/models/dto"
	end2end_utils "annotater/internal/tests/end2end/utils"
	unit_test_utils "annotater/internal/tests/utils"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/signintech/gopdf"
	"github.com/stretchr/testify/suite"
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
func (suite *E2ESuite) SetupSuite() {
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

	err = end2end_utils.TablesMigrate(db)
	suite.Require().NoError(err)

	suite.db = db
	// Initialize httpexpect with a new test.T
	suite.e = *httpexpect.WithConfig(httpexpect.Config{
		Client:   &http.Client{},
		BaseURL:  "http://localhost:8080",
		Reporter: httpexpect.NewAssertReporter(suite.T()),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(suite.T(), true),
		},
	})
	done := make(chan os.Signal, 1)
	suite.done = done
	suite.wg.Add(1)
	go end2end_utils.RunTheApp(suite.db, suite.done, os.Getenv("CONFIG_PATH"), &suite.wg)
}

func (s *E2ESuite) TearDownSuite() {
	ctx := context.Background()
	err := s.pgContainer.Terminate(ctx)
	s.Require().NoError(err)

	s.done <- os.Interrupt
}

func (s *E2ESuite) Test_E2ELoadingDocument() {
	fmt.Print("STARTING THE TEST")
	s.wg.Wait()
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

	jwt := s.e.POST("/user/SignIn").
		WithJSON(reqSignIn).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		NotEmpty().
		ContainsKey("Response").
		Value("jwt").Raw().(string)

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 210, H: 297}}) // A4 size
	bytesPage := pdf.GetBytesPdf()
	r := bytes.NewReader(bytesPage)

	s.e.POST("/document/report").
		WithHeader("Authorization", "Bearer: "+jwt).
		WithMultipart().
		WithFile("file", "filename.pdf", r).
		Expect().
		Status(http.StatusOK)
}

// TestEnd2End runs the test suite
func TestEnd2End(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}
