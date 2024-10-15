//go:build inegration

package integration_tests

import (
	"context"
	"fmt"
	"testing"

	service "annotater/internal/bl/anotattionTypeService"
	repo_adapter "annotater/internal/bl/anotattionTypeService/anottationTypeRepo/anotattionTypeRepoAdapter"
	models_da "annotater/internal/models/modelsDA"
	unit_test_utils "annotater/internal/tests/utils"

	"github.com/stretchr/testify/suite"
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

func (suite *MarkupTypeTestSuite) SetupTest() {
	fmt.Printf("%p", suite)
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
	suite.Require().NoError(err)
	suite.pgContainer = pgContainer

	host, err := pgContainer.Host(ctx)
	suite.Require().NoError(err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	suite.Require().NoError(err)

	dsn := fmt.Sprintf("host=%s port=%s user=test_user password=test_password dbname=test_db sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	suite.Require().NoError(err)

	db.AutoMigrate(&models_da.MarkupType{})
	suite.db = db
}

func (suite *MarkupTypeTestSuite) TestUsecaseAddMarkUpType() {

	anotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(suite.db)
	anotattionTypeService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, anotattionTypeRepo)

	markUpTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
	markUpType := markUpTypeMother.NewDefaultMarkupType()

	gotMarkUpType := models_da.MarkupType{ID: markUpType.ID}
	suite.Require().Error(suite.db.Model(&models_da.MarkupType{}).Where("id = ?", markUpType.ID).Take(&gotMarkUpType).Error)

	err := anotattionTypeService.AddAnottationType(markUpType)
	suite.Require().NoError(err)

	suite.Assert().NoError(suite.db.Model(&models_da.MarkupType{}).Where("id = ?", markUpType.ID).Take(&gotMarkUpType).Error)
	suite.Assert().Equal(models_da.FromDaMarkupType(&gotMarkUpType), *markUpType)
}

func (suite *MarkupTypeTestSuite) TestUsecaseGetMarkUpType() {
	anotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(suite.db)
	anotattionTypeService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, anotattionTypeRepo)

	markUpTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
	markUpType := markUpTypeMother.NewDefaultMarkupType()

	markUpTypeDa := models_da.ToDaMarkupType(*markUpType)
	suite.Require().NoError(suite.db.Create(&markUpTypeDa).Error)

	gotMarkUpType, err := anotattionTypeService.GetAnottationTypeByID(markUpType.ID)
	suite.Require().NoError(err)

	suite.Assert().Equal(markUpType, gotMarkUpType)
}

func (suite *MarkupTypeTestSuite) TestUsecaseDeleteMarkUpType() {
	anotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(suite.db)
	anotattionTypeService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, anotattionTypeRepo)

	markUpTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
	markUpType := markUpTypeMother.NewDefaultMarkupType()

	markUpTypeDa := models_da.ToDaMarkupType(*markUpType)
	suite.Require().NoError(suite.db.Create(&markUpTypeDa).Error)

	err := anotattionTypeService.DeleteAnotattionType(markUpType.ID)
	suite.Require().NoError(err)

	gotMarkUpType := models_da.MarkupType{ID: markUpType.ID}
	suite.Require().Error(suite.db.Model(&models_da.MarkupType{}).Where("id = ?", markUpType.ID).Take(&gotMarkUpType).Error)
}

func (suite *MarkupTypeTestSuite) TearDownTest() {
	suite.db.Migrator().DropTable(&models_da.MarkupType{})

	ctx := context.Background()
	err := suite.pgContainer.Terminate(ctx)
	suite.Require().NoError(err)
}

func TestSuiteMarkupType(t *testing.T) {
	suite.Run(t, new(MarkupTypeTestSuite))
}
