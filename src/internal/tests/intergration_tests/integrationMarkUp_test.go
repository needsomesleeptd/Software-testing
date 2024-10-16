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

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MarkupTestSuite struct {
	suite.Suite
	db          *gorm.DB
	pgContainer testcontainers.Container
}

func (suite *MarkupTestSuite) SetupTest() {
	ctx := context.Background()
	fmt.Print(suite)
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

	host, err := pgContainer.Host(ctx)
	suite.Require().NoError(err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	suite.Require().NoError(err)

	dsn := fmt.Sprintf("host=%s port=%s user=test_user password=test_password dbname=test_db sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{})
	suite.Require().NoError(err)

	db.AutoMigrate(&models_da.Markup{})
	suite.db = db
}

func (suite *MarkupTestSuite) TestUsecaseAddMarkUp() {
	markupMother := unit_test_utils.NewMarkupBuilder()
	markup := markupMother.WithPageData(unit_test_utils.VALID_PNG_BUFFER).
		WithCreatorID(1).
		WithEID(unit_test_utils.TEST_BASIC_ID).
		WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
		Build()

	anotattionRepo := annot_repo_adapter.NewAnotattionRepositoryAdapter(suite.db)
	anotattionService := annot_service.NewAnnotattionService(unit_test_utils.MockLogger, anotattionRepo)

	gotMarkUp := models_da.Markup{ID: markup.ID}
	suite.Require().Error(suite.db.Model(&models_da.Markup{}).Where("id = ?", markup.ID).Take(&gotMarkUp).Error)

	err := anotattionService.AddAnottation(markup)
	suite.Require().NoError(err)

	suite.Assert().NoError(suite.db.Model(&models_da.Markup{}).Where("id = ?", markup.ID).Take(&gotMarkUp).Error)
	markUpNew, err := models_da.FromDaMarkup(&gotMarkUp)
	suite.Require().NoError(err)
	suite.Assert().Equal(markUpNew, *markup)
}

func (suite *MarkupTestSuite) TestUsecaseDeleteMarkUp() {
	markupMother := unit_test_utils.NewMarkupBuilder()
	markup := markupMother.WithPageData(unit_test_utils.VALID_PNG_BUFFER).
		WithCreatorID(1).
		WithEID(unit_test_utils.TEST_BASIC_ID).
		WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
		Build()

	anotattionRepo := annot_repo_adapter.NewAnotattionRepositoryAdapter(suite.db)
	anotattionService := annot_service.NewAnnotattionService(unit_test_utils.MockLogger, anotattionRepo)

	err := anotattionService.AddAnottation(markup)
	suite.Require().NoError(err)

	gotMarkUp := models_da.Markup{ID: markup.ID}
	suite.Require().NoError(suite.db.Model(&models_da.Markup{}).Where("id = ?", markup.ID).Take(&gotMarkUp).Error)

	err = anotattionService.DeleteAnotattion(markup.ID)
	suite.Require().NoError(err)
	suite.Require().Error(suite.db.Model(&models_da.Markup{}).Where("id = ?", markup.ID).Take(&gotMarkUp).Error)
}

func (suite *MarkupTestSuite) TestUsecaseGetMarkUp() {
	markupMother := unit_test_utils.NewMarkupBuilder()
	markup := markupMother.WithPageData(unit_test_utils.VALID_PNG_BUFFER).
		WithCreatorID(1).
		WithEID(unit_test_utils.TEST_BASIC_ID).
		WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
		Build()

	markupDa, err := models_da.ToDaMarkup(*markup)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.db.Create(&markupDa).Error)

	anotattionRepo := annot_repo_adapter.NewAnotattionRepositoryAdapter(suite.db)
	anotattionService := annot_service.NewAnnotattionService(unit_test_utils.MockLogger, anotattionRepo)

	markUp, err := anotattionService.GetAnottationByID(markupDa.ID)
	suite.Require().NoError(err)

	markUpNew, _ := models_da.FromDaMarkup(markupDa)
	suite.Require().Equal(*markUp, markUpNew)
}

func (suite *MarkupTestSuite) TearDownTest() {
	suite.db.Migrator().DropTable(&models_da.Markup{})

	ctx := context.Background()
	err := suite.pgContainer.Terminate(ctx)
	suite.Require().NoError(err)
}

func TestSuiteMarkup(t *testing.T) {
	suite.Run(t, new(MarkupTestSuite))
}
