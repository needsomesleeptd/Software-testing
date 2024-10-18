//go:build integration

package integration_tests

import (
	"context"
	"fmt"
	"testing"

	service "annotater/internal/bl/anotattionTypeService"
	repo_adapter "annotater/internal/bl/anotattionTypeService/anottationTypeRepo/anotattionTypeRepoAdapter"
	models_da "annotater/internal/models/modelsDA"
	unit_test_utils "annotater/internal/tests/utils"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MarkupTypeTestSuite struct {
	suite.Suite
	db          *gorm.DB
	pgContainer testcontainers.Container
}

func (suite *MarkupTypeTestSuite) BeforeEach(t provider.T) {
	ctx := context.Background()

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

	host, err := pgContainer.Host(ctx)
	t.Require().NoError(err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	t.Require().NoError(err)

	dsn := fmt.Sprintf("host=%s port=%s user=test_user password=test_password dbname=test_db sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	t.Require().NoError(err)

	t.Require().NoError(db.AutoMigrate(&models_da.MarkupType{}))
	suite.db = db
}

func (suite *MarkupTypeTestSuite) TearDownTest(t provider.T) {
	// Cleanup steps can be logged to Allure
	t.Log("Cleaning up test environment...")

	suite.db.Migrator().DropTable(&models_da.MarkupType{})

	ctx := context.Background()
	err := suite.pgContainer.Terminate(ctx)
	t.Require().NoError(err)
}

func (suite *MarkupTypeTestSuite) TestUsecaseAddMarkUpType(t provider.T) {

	anotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(suite.db)
	anotattionTypeService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, anotattionTypeRepo)

	markUpTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
	markUpType := markUpTypeMother.NewDefaultMarkupType()

	gotMarkUpType := models_da.MarkupType{ID: markUpType.ID}
	t.Require().Error(suite.db.Model(&models_da.MarkupType{}).Where("id = ?", markUpType.ID).Take(&gotMarkUpType).Error)

	err := anotattionTypeService.AddAnottationType(markUpType)
	t.Require().NoError(err)

	t.Assert().NoError(suite.db.Model(&models_da.MarkupType{}).Where("id = ?", markUpType.ID).Take(&gotMarkUpType).Error)
	t.Assert().Equal(models_da.FromDaMarkupType(&gotMarkUpType), *markUpType)

}

func (suite *MarkupTypeTestSuite) TestUsecaseGetMarkUpType(t provider.T) {

	anotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(suite.db)
	anotattionTypeService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, anotattionTypeRepo)

	markUpTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
	markUpType := markUpTypeMother.NewDefaultMarkupType()

	markUpTypeDa := models_da.ToDaMarkupType(*markUpType)
	t.Require().NoError(suite.db.Create(&markUpTypeDa).Error)

	gotMarkUpType, err := anotattionTypeService.GetAnottationTypeByID(markUpType.ID)
	t.Require().NoError(err)

	t.Assert().Equal(markUpType, gotMarkUpType)
}

func (suite *MarkupTypeTestSuite) TestUsecaseDeleteMarkUpType(t provider.T) {

	anotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(suite.db)
	anotattionTypeService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, anotattionTypeRepo)

	markUpTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
	markUpType := markUpTypeMother.NewDefaultMarkupType()

	markUpTypeDa := models_da.ToDaMarkupType(*markUpType)
	t.Require().NoError(suite.db.Create(&markUpTypeDa).Error)

	err := anotattionTypeService.DeleteAnotattionType(markUpType.ID)
	t.Require().NoError(err)

	gotMarkUpType := models_da.MarkupType{ID: markUpType.ID}
	t.Require().Error(suite.db.Model(&models_da.MarkupType{}).Where("id = ?", markUpType.ID).Take(&gotMarkUpType).Error)

}

func TestSuiteMarkupType(t *testing.T) {
	suite.RunSuite(t, new(MarkupTypeTestSuite))
}