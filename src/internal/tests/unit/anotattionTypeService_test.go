package service_test

import (
	service "annotater/internal/bl/anotattionTypeService"
	mock_repository "annotater/internal/mocks/bl/anotattionTypeService/anottationTypeRepo"
	"annotater/internal/models"
	unit_test_utils "annotater/internal/tests/utils"
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/golang/mock/gomock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type AnnotattionTypeServiceSuite struct {
	suite.Suite
}

func (s *AnnotattionTypeServiceSuite) Test_AnotattionTypeService_AddAnottationType(t provider.T) {
	type fields struct {
		repo *mock_repository.MockIAnotattionTypeRepository
	}
	type args struct {
		anotattionType *models.MarkupType
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		beforeTest func(f *fields)
		wantErr    bool
		err        error
	}{
		{
			name: "Add no err",
			beforeTest: func(f *fields) {
				f.repo.EXPECT().AddAnottationType(unit_test_utils.NewMarkupTypeObjectMother().
					NewDefaultMarkupType()).
					Return(nil)
			},
			wantErr: false,
			err:     nil,
			args:    args{anotattionType: unit_test_utils.NewMarkupTypeObjectMother().NewDefaultMarkupType()},
		},
		{
			name: "Add with err",
			beforeTest: func(f *fields) {
				f.repo.EXPECT().AddAnottationType(unit_test_utils.
					NewMarkupTypeObjectMother().
					NewDefaultMarkupType()).Return(unit_test_utils.ErrEmpty)
			},
			wantErr: true,
			args: args{anotattionType: unit_test_utils.
				NewMarkupTypeObjectMother().
				NewDefaultMarkupType()},
			err: errors.Wrapf(unit_test_utils.ErrEmpty, service.ADDING_ANNOTATTION_ERR_STRF, unit_test_utils.TEST_BASIC_ID),
		},
	}
	t.Title("AddAnotattionType")
	t.Tags("annotattionType")
	for _, tt := range tests {
		//t.Parallel()
		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			ctx := context.TODO()
			ctrl := gomock.NewController(t)
			annotattionMockStorage := mock_repository.NewMockIAnotattionTypeRepository(ctrl)
			if tt.beforeTest != nil {
				tt.beforeTest(&fields{
					repo: annotattionMockStorage,
				})
			}

			annotService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, annotattionMockStorage)
			err := annotService.AddAnottationType(tt.args.anotattionType)

			sCtx.WithNewParameters("ctx", ctx, "err", err)
			if tt.wantErr {
				sCtx.Assert().Error(err)
				sCtx.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				sCtx.Assert().NoError(err)
			}
		})
	}
}

func (s *AnnotattionTypeServiceSuite) Test_AnotattionTypeService_DeleteAnotattionType(t provider.T) {
	type fields struct {
		repo *mock_repository.MockIAnotattionTypeRepository
	}
	type args struct {
		anotattionTypeID uint64
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		beforeTest func(f *fields)
		wantErr    bool
		err        error
	}{
		{
			name: "Delete no err",
			beforeTest: func(f *fields) {
				f.repo.EXPECT().DeleteAnotattionType(unit_test_utils.TEST_BASIC_ID).Return(nil)
			},
			wantErr: false,
			err:     nil,
			args:    args{anotattionTypeID: unit_test_utils.TEST_BASIC_ID},
		},
		{
			name: "Delete with repo err",
			beforeTest: func(f *fields) {
				f.repo.EXPECT().DeleteAnotattionType(unit_test_utils.TEST_BASIC_ID).Return(unit_test_utils.ErrEmpty)
			},
			wantErr: true,
			args:    args{anotattionTypeID: unit_test_utils.TEST_BASIC_ID},
			err:     errors.Wrapf(unit_test_utils.ErrEmpty, service.DELETING_ANNOTATTION_ERR_STRF, unit_test_utils.TEST_BASIC_ID),
		},
	}
	t.Title("DeleteAnotattionType")
	t.Tags("annotattionType")
	for _, tt := range tests {

		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {

			ctx := context.TODO()
			ctrl := gomock.NewController(t)
			annotattionMockStorage := mock_repository.NewMockIAnotattionTypeRepository(ctrl)
			if tt.beforeTest != nil {
				tt.beforeTest(&fields{
					repo: annotattionMockStorage,
				})
			}

			annotService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, annotattionMockStorage)
			err := annotService.DeleteAnotattionType(tt.args.anotattionTypeID)

			sCtx.WithNewParameters("ctx", ctx, "err", err)
			if tt.wantErr {
				sCtx.Assert().Error(err)
				sCtx.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				sCtx.Assert().NoError(err)
			}
		})
	}
}

func (s *AnnotattionTypeServiceSuite) Test_AnotattionTypeService_GetAnottationTypeByID(t provider.T) {
	type fields struct {
		repo *mock_repository.MockIAnotattionTypeRepository
	}
	type args struct {
		id uint64
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		beforeTest func(f *fields)
		wantErr    bool
		err        error
		want       *models.MarkupType
	}{
		{
			name: "Get no err",
			beforeTest: func(f *fields) {
				f.repo.EXPECT().GetAnottationTypeByID(unit_test_utils.TEST_BASIC_ID).Return(&models.MarkupType{ID: unit_test_utils.TEST_BASIC_ID}, nil)
			},
			wantErr: false,
			err:     nil,
			args:    args{id: unit_test_utils.TEST_BASIC_ID},
			want:    &models.MarkupType{ID: unit_test_utils.TEST_BASIC_ID},
		},
		{
			name: "Get with repo err",
			beforeTest: func(f *fields) {
				f.repo.EXPECT().GetAnottationTypeByID(unit_test_utils.TEST_BASIC_ID).Return(nil, unit_test_utils.ErrEmpty)
			},
			wantErr: true,
			args:    args{id: unit_test_utils.TEST_BASIC_ID},
			err:     errors.Wrapf(unit_test_utils.ErrEmpty, service.GETTING_ANNOTATTION_STR_ERR_STRF, unit_test_utils.TEST_BASIC_ID),
			want:    nil,
		},
	}
	t.Title("GetAnottationTypeByID")
	t.Tags("annotattionType")
	for _, tt := range tests {
		//t.Parallel()
		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			ctx := context.TODO()
			ctrl := gomock.NewController(t)
			annotattionMockStorage := mock_repository.NewMockIAnotattionTypeRepository(ctrl)
			if tt.beforeTest != nil {
				tt.beforeTest(&fields{
					repo: annotattionMockStorage,
				})
			}

			annotService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, annotattionMockStorage)
			res, err := annotService.GetAnottationTypeByID(tt.args.id)

			sCtx.WithNewParameters("ctx", ctx, "err", err)
			if tt.wantErr {
				sCtx.Assert().Error(err)
				sCtx.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				sCtx.Assert().NoError(err)
				sCtx.Assert().Equal(tt.want, res)
			}
		})
	}
}

func TestAnotattionTypeSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(AnnotattionTypeServiceSuite))
}
