//go:build integration

package integration_tests

import (
	annot_service "annotater/internal/bl/annotationService"
	annot_repo_adapter "annotater/internal/bl/annotationService/annotattionRepo/anotattionRepoAdapter"
	models_da "annotater/internal/models/modelsDA"
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

type MarkupTestSuite struct {
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
	err = db.AutoMigrate(&models_da.Markup{})
	t.Require().NoError(err) // Ensure migration succeeds

	return pgContainer, db
}

func destroyContainer(t provider.T, pgContainer testcontainers.Container) {
	// Cleanup the container after each test
	ctx := context.Background()
	err := pgContainer.Terminate(ctx)
	t.Require().NoError(err)
}

func (suite *MarkupTestSuite) TestUsecaseAddMarkUp(t provider.T) {

	container, db := createDBInContainer(t)
	defer destroyContainer(t, container)
	t.Require().NotNil(db)

	markupMother := unit_test_utils.NewMarkupBuilder()
	markup := markupMother.WithPageData(unit_test_utils.VALID_PNG_BUFFER).
		WithCreatorID(1).
		WithEID(unit_test_utils.TEST_BASIC_ID).
		WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
		Build()

	anotattionRepo := annot_repo_adapter.NewAnotattionRepositoryAdapter(db)
	anotattionService := annot_service.NewAnnotattionService(unit_test_utils.MockLogger, anotattionRepo)

	gotMarkUp := models_da.Markup{ID: markup.ID}
	t.Require().Error(db.Model(&models_da.Markup{}).Where("id = ?", markup.ID).Take(&gotMarkUp).Error)

	err := anotattionService.AddAnottation(markup)
	t.Require().NoError(err)

	t.Assert().NoError(db.Model(&models_da.Markup{}).Where("id = ?", markup.ID).Take(&gotMarkUp).Error)
	markUpNew, err := models_da.FromDaMarkup(&gotMarkUp)
	t.Require().NoError(err)
	t.Assert().Equal(markUpNew, *markup)

}

func (suite *MarkupTestSuite) TestUsecaseDeleteMarkUp(t provider.T) {

	container, db := createDBInContainer(t)
	defer destroyContainer(t, container)
	t.Require().NotNil(db)

	markupMother := unit_test_utils.NewMarkupBuilder()
	markup := markupMother.WithPageData(unit_test_utils.VALID_PNG_BUFFER).
		WithCreatorID(1).
		WithEID(unit_test_utils.TEST_BASIC_ID).
		WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
		Build()

	anotattionRepo := annot_repo_adapter.NewAnotattionRepositoryAdapter(db)
	anotattionService := annot_service.NewAnnotattionService(unit_test_utils.MockLogger, anotattionRepo)

	markupDa, err := models_da.ToDaMarkup(*markup)
	t.Require().NoError(err)

	t.Require().NoError(db.Create(&markupDa).Error)

	gotMarkUp := models_da.Markup{ID: markup.ID}
	t.Require().NoError(db.Model(&models_da.Markup{}).Where("id = ?", markup.ID).Take(&gotMarkUp).Error)

	err = anotattionService.DeleteAnotattion(markup.ID)
	t.Require().NoError(err)

	t.Require().Error(db.Model(&models_da.Markup{}).Where("id = ?", markup.ID).Take(&gotMarkUp).Error)

}

func (suite *MarkupTestSuite) TestUsecaseGetMarkUp(t provider.T) {

	container, db := createDBInContainer(t)
	defer destroyContainer(t, container)
	t.Require().NotNil(db)

	markupMother := unit_test_utils.NewMarkupBuilder()
	markup := markupMother.WithPageData(unit_test_utils.VALID_PNG_BUFFER).
		WithCreatorID(1).
		WithEID(unit_test_utils.TEST_BASIC_ID).
		WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
		Build()

	markupDa, err := models_da.ToDaMarkup(*markup)
	t.Require().NoError(err)
	t.Require().NoError(db.Create(&markupDa).Error)

	anotattionRepo := annot_repo_adapter.NewAnotattionRepositoryAdapter(db)
	anotattionService := annot_service.NewAnnotattionService(unit_test_utils.MockLogger, anotattionRepo)

	markUp, err := anotattionService.GetAnottationByID(markupDa.ID)
	t.Require().NoError(err)

	markUpNew, err := models_da.FromDaMarkup(markupDa)
	t.Require().NoError(err)

	t.Require().Equal(*markUp, markUpNew)

}

func TestSuiteMarkup(t *testing.T) {
	suite.RunSuite(t, new(MarkupTestSuite))
}
