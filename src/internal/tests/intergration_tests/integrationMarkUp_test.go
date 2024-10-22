//go:build integration

package integration_tests

import (
	annot_service "annotater/internal/bl/annotationService"
	annot_repo_adapter "annotater/internal/bl/annotationService/annotattionRepo/anotattionRepoAdapter"
	models_da "annotater/internal/models/modelsDA"
	integration_utils "annotater/internal/tests/intergration_tests/utils"
	unit_test_utils "annotater/internal/tests/utils"
	"os"
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type MarkupTestSuite struct {
	suite.Suite
}

func (suite *MarkupTestSuite) TestUsecaseAddMarkUp(t provider.T) {
	if os.Getenv("UNIT_FAILED") != "" {
		t.Skip("Unit test failed, skipping")
	}

	container, db := integration_utils.CreateDBInContainer(t)
	defer integration_utils.DestroyContainer(t, container)
	t.Require().NotNil(db)
	err := db.AutoMigrate(&models_da.Markup{})
	t.Require().NoError(err)

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

	err = anotattionService.AddAnottation(markup)
	t.Require().NoError(err)

	t.Assert().NoError(db.Model(&models_da.Markup{}).Where("id = ?", markup.ID).Take(&gotMarkUp).Error)
	markUpNew, err := models_da.FromDaMarkup(&gotMarkUp)
	t.Require().NoError(err)
	t.Assert().Equal(markUpNew, *markup)

}

func (suite *MarkupTestSuite) TestUsecaseDeleteMarkUp(t provider.T) {
	if os.Getenv("UNIT_FAILED") != "" {
		t.Skip("Unit test failed, skipping")
	}
	container, db := integration_utils.CreateDBInContainer(t)
	defer integration_utils.DestroyContainer(t, container)
	t.Require().NotNil(db)
	err := db.AutoMigrate(&models_da.Markup{})
	t.Require().NoError(err)

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
	if os.Getenv("UNIT_FAILED") != "" {
		t.Skip("Unit test failed, skipping")
	}

	container, db := integration_utils.CreateDBInContainer(t)
	defer integration_utils.DestroyContainer(t, container)
	t.Require().NotNil(db)
	err := db.AutoMigrate(&models_da.Markup{})
	t.Require().NoError(err)

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
