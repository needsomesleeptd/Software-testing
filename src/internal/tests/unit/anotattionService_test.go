package service_test

import (
	service "annotater/internal/bl/annotationService"
	mock_repository "annotater/internal/mocks/bl/annotationService/annotattionRepo"
	"annotater/internal/models"
	unit_test_utils "annotater/internal/tests/utils"
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/golang/mock/gomock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type AnnotattionServiceSuite struct {
	suite.Suite
}

func (s *AnnotattionServiceSuite) Test_AnnotattionService_AddAnnotation(t provider.T) {
	tests := []struct {
		name      string
		markup    *models.Markup
		expectErr bool
	}{
		{
			name: "[AddAnnotation] Valid annotation",
			markup: unit_test_utils.NewMarkupBuilder().
				WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
				WithPageData(unit_test_utils.VALID_PNG_BUFFER).
				Build(),
			expectErr: false,
		},
		{
			name: "[AddAnnotation] Invalid markup BBs",
			markup: unit_test_utils.NewMarkupBuilder().
				WithErrorBB(unit_test_utils.INVALID_BBS_PARAMS).
				WithPageData(unit_test_utils.VALID_PNG_BUFFER).
				Build(),
			expectErr: true,
		},
		{
			name: "[AddAnnotation] Invalid page",
			markup: unit_test_utils.NewMarkupBuilder().
				WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
				WithPageData(unit_test_utils.INVALD_PNG_BUFFER).
				Build(),
			expectErr: true,
		},
	}
	t.Title("AddAnnotation")
	t.Tags("annotattionService")
	for _, tt := range tests {

		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			ctx := context.TODO()
			ctrl := gomock.NewController(t)

			annotattionMockStorage := mock_repository.NewMockIAnotattionRepository(ctrl)

			if !tt.expectErr {
				annotattionMockStorage.EXPECT().AddAnottation(tt.markup)
			}

			annotService := service.NewAnnotattionService(unit_test_utils.MockLogger, annotattionMockStorage)
			err := annotService.AddAnottation(tt.markup)

			sCtx.WithNewParameters("ctx", ctx, "markup", tt.markup)

			if tt.expectErr {
				sCtx.Assert().Error(err)
			} else {
				sCtx.Assert().NoError(err)
			}
		})
	}
}

func (s *AnnotattionServiceSuite) Test_areBBsValid(t provider.T) {
	tests := []struct {
		name    string
		bbs     []float32
		want    bool
		wantErr string
	}{
		{
			name:    "[AreBBsValid] Valid slice",
			bbs:     []float32{1.0, 0.0, 0.0, 1.0},
			want:    true,
			wantErr: "",
		},
		{
			name:    "[AreBBsValid] Invalid neg slice",
			bbs:     []float32{-1.0, 0.0, 0.0, 1.0},
			want:    false,
			wantErr: "",
		},
		{
			name:    "[AreBBsValid] Invalid bigger than 1 slice",
			bbs:     []float32{1.0, 0.0, 0.0, 1.1},
			want:    false,
			wantErr: "",
		},
	}
	t.Title("areBBsValid")
	t.Tags("annotattionService")
	for _, tt := range tests {

		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			sCtx.WithNewParameters("bbs", tt.bbs)

			result := service.AreBBsValid(tt.bbs)
			sCtx.Assert().Equal(tt.want, result)
		})
	}
}

func (s *AnnotattionServiceSuite) TestAnotattionService_DeleteAnotattion(t provider.T) {
	tests := []struct {
		name      string
		id        uint64
		returnErr error
		wantErr   bool
	}{
		{
			name:      "[DeleteAnotattion] Delete no error",
			id:        unit_test_utils.TEST_BASIC_ID,
			returnErr: nil,
			wantErr:   false,
		},
		{
			name:      "[DeleteAnotattion] Delete with repository error",
			id:        unit_test_utils.TEST_BASIC_ID,
			returnErr: errors.New(""),
			wantErr:   true,
		},
	}

	t.Title("DeleteAnotattion")
	t.Tags("annotattionService")
	for _, tt := range tests {

		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			ctx := context.TODO()
			ctrl := gomock.NewController(t)

			annotattionMockStorage := mock_repository.NewMockIAnotattionRepository(ctrl)
			annotattionMockStorage.EXPECT().DeleteAnotattion(tt.id).Return(tt.returnErr)

			annotService := service.NewAnnotattionService(unit_test_utils.MockLogger, annotattionMockStorage)
			err := annotService.DeleteAnotattion(tt.id)

			sCtx.WithNewParameters("ctx", ctx, "id", tt.id)
			if tt.wantErr {
				sCtx.Assert().Error(err)
			} else {
				sCtx.Assert().NoError(err)
			}
		})
	}
}

func (s *AnnotattionServiceSuite) TestAnotattionService_GetAnottationByID(t provider.T) {
	tests := []struct {
		beforeTest func(finRepo mock_repository.MockIAnotattionRepository)
		name       string
		id         uint64
		wantErr    bool
		err        error
		want       *models.Markup
	}{
		{
			beforeTest: func(finRepo mock_repository.MockIAnotattionRepository) {
				finRepo.EXPECT().GetAnottationByID(unit_test_utils.TEST_BASIC_ID).Return(unit_test_utils.VALID_MARKUP, nil)
			},
			name:    "[GetAnottationByID] Get without repository error",
			id:      unit_test_utils.TEST_BASIC_ID,
			wantErr: false,
			err:     nil,
			want:    unit_test_utils.VALID_MARKUP,
		},
		{
			beforeTest: func(finRepo mock_repository.MockIAnotattionRepository) {
				finRepo.EXPECT().GetAnottationByID(unit_test_utils.TEST_BASIC_ID).Return(nil, errors.New(""))
			},
			name:    "[GetAnottationByID] Get with repository error",
			id:      unit_test_utils.TEST_BASIC_ID,
			wantErr: true,
			err:     errors.Wrapf(errors.New(""), service.GETTING_ANNOT_ERR_STRF, unit_test_utils.TEST_BASIC_ID),
			want:    nil,
		},
	}
	t.Title("GetAnottationByID")
	t.Tags("annotattionService")
	for _, tt := range tests {
		//t.Parallel()
		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			ctx := context.TODO()
			ctrl := gomock.NewController(t)
			annotattionMockStorage := mock_repository.NewMockIAnotattionRepository(ctrl)
			if tt.beforeTest != nil {
				tt.beforeTest(*annotattionMockStorage)
			}

			annotService := service.NewAnnotattionService(unit_test_utils.MockLogger, annotattionMockStorage)
			markup, err := annotService.GetAnottationByID(tt.id)

			sCtx.WithNewParameters("ctx", ctx, "id", tt.id)
			if tt.wantErr {
				sCtx.Assert().Error(err)
				sCtx.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				sCtx.Assert().NoError(err)
				sCtx.Assert().Equal(markup, tt.want)
			}
		})
	}
}

func (s *AnnotattionServiceSuite) Test_checkPngFile(t provider.T) {
	tests := []struct {
		name      string
		pngFile   []byte
		expectErr bool
	}{
		{
			name:      "[CheckPngFile] Valid PNG",
			pngFile:   unit_test_utils.VALID_PNG_BUFFER,
			expectErr: false,
		},
		{
			name:      "[CheckPngFile] Invalid PNG",
			pngFile:   unit_test_utils.INVALD_PNG_BUFFER,
			expectErr: true,
		},
	}
	t.Title("CheckPngFile")
	t.Tags("annotattionService")
	for _, tt := range tests {
		// t.Parallel()

		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			ctx := context.TODO()

			sCtx.WithNewParameters("ctx", ctx, "pngFile", tt.pngFile)
			err := service.CheckPngFile(tt.pngFile)

			if tt.expectErr {
				sCtx.Assert().Error(err)
			} else {
				sCtx.Assert().NoError(err)
			}
		})
	}
}
func TestAnotattionSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(AnnotattionServiceSuite))
}
