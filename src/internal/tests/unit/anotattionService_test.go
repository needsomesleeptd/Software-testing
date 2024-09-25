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
	t.Title("[AddAnnotation] Valid annotation")
	t.Tags("annotattion")
	//t.Parallel()
	t.WithNewStep("Valid annotation", func(sCtx provider.StepCtx) {
		ctx := context.TODO()
		ctrl := gomock.NewController(t)
		markup := unit_test_utils.NewMarkupBuilder().
			WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
			WithPageData(unit_test_utils.VALID_PNG_BUFFER).
			Build()

		annotattionMockStorage := mock_repository.NewMockIAnotattionRepository(ctrl)
		annotattionMockStorage.EXPECT().AddAnottation(markup)

		annotService := service.NewAnnotattionService(unit_test_utils.MockLogger, annotattionMockStorage)
		err := annotService.AddAnottation(markup)

		sCtx.WithNewParameters("ctx", ctx, "markup", markup)
		sCtx.Assert().NoError(err)
	})

	t.Title("[AddAnnotation] Invalid markup BBs")
	t.Tags("annotattion")
	//	t.Parallel()
	t.WithNewStep("Invalid markup BBs", func(sCtx provider.StepCtx) {
		ctx := context.TODO()
		ctrl := gomock.NewController(t)
		markup := unit_test_utils.NewMarkupBuilder().
			WithErrorBB(unit_test_utils.INVALID_BBS_PARAMS).
			WithPageData(unit_test_utils.VALID_PNG_BUFFER).
			Build()

		annotattionMockStorage := mock_repository.NewMockIAnotattionRepository(ctrl)
		//annotattionMockStorage.EXPECT().AddAnottation(markup)

		annotService := service.NewAnnotattionService(unit_test_utils.MockLogger, annotattionMockStorage)
		err := annotService.AddAnottation(markup)
		sCtx.WithNewParameters("ctx", ctx, "markup", markup, "error", err)

		sCtx.Assert().Error(err)
	})

	t.Title("[AddAnnotation] Invalid page")
	t.Tags("annotattion")
	//t.Parallel()
	t.WithNewStep("Invalid page", func(sCtx provider.StepCtx) {
		ctx := context.TODO()
		ctrl := gomock.NewController(t)
		markup := unit_test_utils.NewMarkupBuilder().
			WithErrorBB(unit_test_utils.VALID_BBS_PARAMS).
			WithPageData(unit_test_utils.INVALD_PNG_BUFFER).
			Build()

		annotattionMockStorage := mock_repository.NewMockIAnotattionRepository(ctrl)
		//		annotattionMockStorage.EXPECT().AddAnottation(markup)

		annotService := service.NewAnnotattionService(unit_test_utils.MockLogger, annotattionMockStorage)
		err := annotService.AddAnottation(markup)

		sCtx.WithNewParameters("ctx", ctx, "markup", markup, "err", err)

		sCtx.Assert().Error(err)
	})
}

func (s *AnnotattionServiceSuite) Test_areBBsValid(t provider.T) {
	t.Title("[AreBBsValid] Valid slice")
	t.Tags("annotattion")
	//t.Parallel()
	t.WithNewStep("Valid slice", func(sCtx provider.StepCtx) {

		bbs := []float32{1.0, 0.0, 0.0, 1.0}
		sCtx.WithNewParameters("bbs", bbs)
		sCtx.Assert().True(service.AreBBsValid(bbs))
	})

	t.Title("[AreBBsValid] Invalid neg slice")
	t.Tags("annotattion")
	//	t.Parallel()
	t.WithNewStep("Invalid neg slice", func(sCtx provider.StepCtx) {

		bbs := []float32{-1.0, 0.0, 0.0, 1.0}
		sCtx.WithNewParameters("bbs", bbs)
		sCtx.Assert().False(service.AreBBsValid(bbs))
	})

	t.Title("[AreBBsValid] Invalid bigger than 1 slice")
	t.Tags("annotattion")
	//t.Parallel()
	t.WithNewStep("Invalid bigger than 1 slice", func(sCtx provider.StepCtx) {
		bbs := []float32{1.0, 0.0, 0.0, 1.1}

		sCtx.WithNewParameters("bbs", bbs)
		sCtx.Assert().False(service.AreBBsValid(bbs))
	})
}

func (s *AnnotattionServiceSuite) TestAnotattionService_DeleteAnotattion(t provider.T) {
	t.Title("[DeleteAnotattion] Delete no error")
	t.Tags("annotattion")
	//t.Parallel()
	t.WithNewStep("Delete no error", func(sCtx provider.StepCtx) {
		ctx := context.TODO()
		ctrl := gomock.NewController(t)
		id := unit_test_utils.TEST_BASIC_ID

		annotattionMockStorage := mock_repository.NewMockIAnotattionRepository(ctrl)
		annotattionMockStorage.EXPECT().DeleteAnotattion(id).Return(nil)

		annotService := service.NewAnnotattionService(unit_test_utils.MockLogger, annotattionMockStorage)
		err := annotService.DeleteAnotattion(id)

		sCtx.WithNewParameters("ctx", ctx, "id", id)
		sCtx.Assert().NoError(err)
	})

	t.Title("[DeleteAnotattion] Delete with repository error")
	t.Tags("annotattion")
	//	t.Parallel()
	t.WithNewStep("Delete with repository error", func(sCtx provider.StepCtx) {
		ctx := context.TODO()
		ctrl := gomock.NewController(t)
		id := unit_test_utils.TEST_BASIC_ID

		annotattionMockStorage := mock_repository.NewMockIAnotattionRepository(ctrl)
		annotattionMockStorage.EXPECT().DeleteAnotattion(id).Return(errors.New(""))

		annotService := service.NewAnnotattionService(unit_test_utils.MockLogger, annotattionMockStorage)
		err := annotService.DeleteAnotattion(id)

		sCtx.WithNewParameters("ctx", ctx, "id", id, "error", err)
		sCtx.Assert().Error(err)
	})
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

	for _, tt := range tests {
		t.Title(tt.name)
		t.Tags("annotattion")
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
	t.Title("[CheckPngFile] Valid PNG")
	t.Tags("annotattion")
	//t.Parallel()
	t.WithNewStep("Valid PNG", func(sCtx provider.StepCtx) {
		ctx := context.TODO()
		pngFile := unit_test_utils.VALID_PNG_BUFFER

		sCtx.WithNewParameters("ctx", ctx, "pngFile", pngFile)
		err := service.CheckPngFile(pngFile)
		sCtx.Assert().NoError(err)
	})

	t.Title("[CheckPngFile] Invalid PNG")
	t.Tags("annotattion")
	//t.Parallel()
	t.WithNewStep("Invalid PNG", func(sCtx provider.StepCtx) {
		ctx := context.TODO()
		pngFile := unit_test_utils.INVALD_PNG_BUFFER

		sCtx.WithNewParameters("ctx", ctx, "pngFile", pngFile)
		err := service.CheckPngFile(pngFile)
		sCtx.Assert().Error(err)
	})
}

func TestAnotattionSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(AnnotattionServiceSuite))
}
