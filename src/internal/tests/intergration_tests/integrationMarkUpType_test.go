//go:build integration

package integration_tests

import (
	"os"
	"testing"

	service "annotater/internal/bl/anotattionTypeService"
	repo_adapter "annotater/internal/bl/anotattionTypeService/anottationTypeRepo/anotattionTypeRepoAdapter"
	models_da "annotater/internal/models/modelsDA"
	integration_utils "annotater/internal/tests/intergration_tests/utils"
	unit_test_utils "annotater/internal/tests/utils"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type MarkupTypeTestSuite struct {
	suite.Suite
}

func (suite *MarkupTypeTestSuite) TestUsecaseAddMarkUpType(t provider.T) {
	if os.Getenv("UNIT_FAILED") != "" {
		t.Skip("Unit test failed, skipping")
	}
	container, db := integration_utils.CreateDBInContainer(t)
	defer integration_utils.DestroyContainer(t, container)
	t.Require().NotNil(db)
	err := db.AutoMigrate(&models_da.MarkupType{})
	t.Require().NoError(err)

	anotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(db)
	anotattionTypeService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, anotattionTypeRepo)

	markUpTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
	markUpType := markUpTypeMother.NewDefaultMarkupType()

	gotMarkUpType := models_da.MarkupType{ID: markUpType.ID}
	t.Require().Error(db.Model(&models_da.MarkupType{}).Where("id = ?", markUpType.ID).Take(&gotMarkUpType).Error)

	err = anotattionTypeService.AddAnottationType(markUpType)
	t.Require().NoError(err)

	t.Assert().NoError(db.Model(&models_da.MarkupType{}).Where("id = ?", markUpType.ID).Take(&gotMarkUpType).Error)
	t.Assert().Equal(models_da.FromDaMarkupType(&gotMarkUpType), *markUpType)

}

func (suite *MarkupTypeTestSuite) TestUsecaseGetMarkUpType(t provider.T) {
	if os.Getenv("UNIT_FAILED") != "" {
		t.XSkip()
	}
	container, db := integration_utils.CreateDBInContainer(t)
	defer integration_utils.DestroyContainer(t, container)
	t.Require().NotNil(db)
	err := db.AutoMigrate(&models_da.MarkupType{})
	t.Require().NoError(err)

	anotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(db)
	anotattionTypeService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, anotattionTypeRepo)

	markUpTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
	markUpType := markUpTypeMother.NewDefaultMarkupType()

	markUpTypeDa := models_da.ToDaMarkupType(*markUpType)
	t.Require().NoError(db.Create(&markUpTypeDa).Error)

	gotMarkUpType, err := anotattionTypeService.GetAnottationTypeByID(markUpType.ID)
	t.Require().NoError(err)

	t.Assert().Equal(markUpType, gotMarkUpType)
}

func (suite *MarkupTypeTestSuite) TestUsecaseDeleteMarkUpType(t provider.T) {
	if os.Getenv("UNIT_FAILED") != "" {
		t.Skip("Unit test failed, skipping")
	}
	container, db := integration_utils.CreateDBInContainer(t)
	defer integration_utils.DestroyContainer(t, container)
	t.Require().NotNil(db)
	err := db.AutoMigrate(&models_da.MarkupType{})
	t.Require().NoError(err)

	anotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(db)
	anotattionTypeService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, anotattionTypeRepo)

	markUpTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
	markUpType := markUpTypeMother.NewDefaultMarkupType()

	markUpTypeDa := models_da.ToDaMarkupType(*markUpType)
	t.Require().NoError(db.Create(&markUpTypeDa).Error)

	err = anotattionTypeService.DeleteAnotattionType(markUpType.ID)
	t.Require().NoError(err)

	gotMarkUpType := models_da.MarkupType{ID: markUpType.ID}
	t.Require().Error(db.Model(&models_da.MarkupType{}).Where("id = ?", markUpType.ID).Take(&gotMarkUpType).Error)

}

func TestSuiteMarkupType(t *testing.T) {
	suite.RunSuite(t, new(MarkupTypeTestSuite))
}
