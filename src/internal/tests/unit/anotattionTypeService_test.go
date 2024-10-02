package service_test

import (
	service "annotater/internal/bl/anotattionTypeService"
	repo_adapter "annotater/internal/bl/anotattionTypeService/anottationTypeRepo/anotattionTypeRepoAdapter"
	mock_repository "annotater/internal/mocks/bl/anotattionTypeService/anottationTypeRepo"
	"annotater/internal/models"
	unit_test_utils "annotater/internal/tests/utils"
	unit_test_mappers "annotater/internal/tests/utils/da"
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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

func (s *AnnotattionTypeServiceSuite) TestAnotattionTypeService_AddAnottationType_Classic(t provider.T) {
	type args struct {
		anotattionType *models.MarkupType
	}
	objectMother := unit_test_utils.NewMarkupTypeObjectMother()
	tests := []struct {
		name      string
		args      args
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   bool
		err       error
	}{
		{
			name: "Add no error",
			setupMock: func(mock sqlmock.Sqlmock) {
				markUpType := objectMother.NewDefaultMarkupType()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "markup_types" ("description","creator_id","class_name","id") VALUES ($1,$2,$3,$4) RETURNING "id"`).
					WithArgs(markUpType.Description, markUpType.CreatorID, markUpType.ClassName, markUpType.ID).
					WillReturnRows(
						sqlmock.NewRows([]string{"id"}).AddRow(markUpType.ID))
				mock.ExpectCommit()
			},
			wantErr: false,
			err:     nil,
			args:    args{anotattionType: objectMother.NewDefaultMarkupType()},
		},
		{
			name: "Add with error",
			setupMock: func(mock sqlmock.Sqlmock) {
				markUpType := objectMother.NewDefaultMarkupType()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "markup_types" ("description","creator_id","class_name","id") VALUES ($1,$2,$3,$4) RETURNING "id"`).
					WithArgs(markUpType.Description, markUpType.CreatorID, markUpType.ClassName, markUpType.ID).
					WillReturnError(gorm.ErrDuplicatedKey)
				mock.ExpectRollback()
			},
			wantErr: true,
			args:    args{anotattionType: unit_test_utils.NewMarkupTypeObjectMother().NewDefaultMarkupType()},
			err:     errors.Wrapf(models.ErrDuplicateMarkupType, service.ADDING_ANNOTATTION_ERR_STRF, unit_test_utils.TEST_BASIC_ID),
		},
	}

	t.Title("AddAnotattionTypeClassic")
	t.Tags("annotattionType")

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			ctx := context.TODO()

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			// Setup the mock expectations
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			annotattionMockStorage := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)
			annotService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, annotattionMockStorage)
			err = annotService.AddAnottationType(tt.args.anotattionType)

			sCtx.WithNewParameters("ctx", ctx, "err", err)

			if tt.wantErr {
				sCtx.Assert().Error(err)
				sCtx.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				sCtx.Assert().NoError(err)
			}

			// Ensure all expectations were met
			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}

func (s *AnnotattionTypeServiceSuite) TestAnotattionTypeService_DeleteAnotattionType_Classic(t provider.T) {
	type args struct {
		anotattionTypeID uint64
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   bool
		err       error
	}{
		{
			name: "Delete no error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "markup_types" WHERE id = $1`).
					WithArgs(unit_test_utils.TEST_BASIC_ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
			err:     nil,
			args:    args{anotattionTypeID: unit_test_utils.TEST_BASIC_ID},
		},
		{
			name: "Delete with repository error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "markup_types" WHERE id = $1`).
					WithArgs(unit_test_utils.TEST_BASIC_ID).
					WillReturnError(unit_test_utils.ErrEmpty)
				mock.ExpectRollback()
			},
			wantErr: true,
			args:    args{anotattionTypeID: unit_test_utils.TEST_BASIC_ID},
			err: errors.Wrapf(errors.Wrap(unit_test_utils.ErrEmpty, "Error in deleting anotattion type db"),
				service.DELETING_ANNOTATTION_ERR_STRF, unit_test_utils.TEST_BASIC_ID),
		},
	}

	t.Title("DeleteAnotattionTypeClassic")
	t.Tags("annotattionType")

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			ctx := context.TODO()

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			annotattionMockStorage := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)

			// Setup the mock expectations
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			annotService := service.NewAnotattionTypeService(unit_test_utils.MockLogger, annotattionMockStorage)
			err = annotService.DeleteAnotattionType(tt.args.anotattionTypeID)

			sCtx.WithNewParameters("ctx", ctx, "err", err)

			if tt.wantErr {
				sCtx.Assert().Error(err)
				sCtx.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				sCtx.Assert().NoError(err)
			}

			// Ensure all expectations were met
			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
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

func (s *AnnotattionTypeServiceSuite) TestAnotattionTypeService_GetAnottationTypeByID_Classic(t provider.T) {

	type args struct {
		id uint64
	}

	objectMother := unit_test_utils.NewMarkupTypeObjectMother()
	tests := []struct {
		name      string
		args      args
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   bool
		err       error
		want      *models.MarkupType
	}{
		{
			name: "Get no error",
			setupMock: func(mock sqlmock.Sqlmock) {
				markupType := objectMother.NewMarkupTypeWithID(unit_test_utils.TEST_BASIC_ID)
				mock.ExpectQuery("SELECT * FROM \"markup_types\" WHERE id = $1 AND \"markup_types\".\"id\" = $2 ORDER BY \"markup_types\".\"id\" LIMIT $3").
					WithArgs(unit_test_utils.TEST_BASIC_ID, unit_test_utils.TEST_BASIC_ID, 1).
					WillReturnRows(unit_test_mappers.MapMarkupTypes(markupType))
			},
			wantErr: false,
			err:     nil,
			args:    args{id: unit_test_utils.TEST_BASIC_ID},
			want:    objectMother.NewMarkupTypeWithID(unit_test_utils.TEST_BASIC_ID),
		},
		{
			name: "Get with repository error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM \"markup_types\" WHERE id = $1 AND \"markup_types\".\"id\" = $2 ORDER BY \"markup_types\".\"id\" LIMIT $3").
					WithArgs(unit_test_utils.TEST_BASIC_ID, unit_test_utils.TEST_BASIC_ID, 1).
					WillReturnError(unit_test_utils.ErrEmpty)
			},
			wantErr: true,
			args:    args{id: unit_test_utils.TEST_BASIC_ID},
			err: errors.Wrapf(
				errors.Wrap(unit_test_utils.ErrEmpty, "Error in getting anotattion type db"),
				service.GETTING_ANNOTATTION_STR_ERR_STRF,
				unit_test_utils.TEST_BASIC_ID),
			want: nil,
		},
	}

	t.Title("GetAnottationTypeByIDClassic")
	t.Tags("annotattionType")

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(sCtx provider.StepCtx) {
			ctx := context.TODO()

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			annotattionMockStorage := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)

			// Setup the mock expectations
			if tt.setupMock != nil {
				tt.setupMock(mock)
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

func (s *AnnotattionTypeServiceSuite) TestAnotattionTypeService_GetAllAnotattionTypes(t provider.T) {

	tests := []struct {
		name           string
		setupMock      func(mock sqlmock.Sqlmock)
		expectedLength int
		expectError    bool
	}{
		{
			name: "Success - Two Types",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "creator_id"}).
					AddRow(1, 1).
					AddRow(2, 2)

				mock.ExpectQuery(`SELECT * FROM "markup_types"`).
					WillReturnRows(rows)
			},
			expectedLength: 2,
			expectError:    false,
		},
		{
			name: "Success - No Types",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "creator_id"})

				mock.ExpectQuery(`SELECT * FROM "markup_types"`).
					WillReturnRows(rows)
			},
			expectedLength: 0,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			annotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)

			// Setup the mock expectations
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			types, err := annotattionTypeRepo.GetAllAnottationTypes()
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			if len(types) != tt.expectedLength {
				t.Errorf("Expected %d annotation types, got %d", tt.expectedLength, len(types))
			}

			mock.ExpectationsWereMet()
		})
	}
}

func (s *AnnotattionTypeServiceSuite) TestAnotattionTypeService_GetAnottationTypesByUserID(t provider.T) {
	tests := []struct {
		name           string
		userID         uint64
		setupMock      func(mock sqlmock.Sqlmock)
		expectedLength int
		expectError    bool
	}{
		{
			name:   "Success - Two Types for User",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "creator_id"}).
					AddRow(1, 1).
					AddRow(2, 1)

				mock.ExpectQuery(`SELECT * FROM "markup_types" WHERE creator_id = $1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedLength: 2,
			expectError:    false,
		},
		{
			name:   "Success - No Types for User",
			userID: 2,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "creator_id"})

				mock.ExpectQuery(`SELECT * FROM "markup_types" WHERE creator_id = $1`).
					WithArgs(2).
					WillReturnRows(rows)
			},
			expectedLength: 0,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			annotattionTypeRepo := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)

			// Setup the mock expectations
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			types, err := annotattionTypeRepo.GetAnottationTypesByUserID(tt.userID)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			if len(types) != tt.expectedLength {
				t.Errorf("Expected %d annotation types, got %d", tt.expectedLength, len(types))
			}

			mock.ExpectationsWereMet()
		})
	}
}

func TestAnotattionTypeSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(AnnotattionTypeServiceSuite))
}
