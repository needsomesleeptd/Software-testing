package service_test

import (
	repo_adapter "annotater/internal/bl/annotationService/annotattionRepo/anotattionRepoAdapter"
	"annotater/internal/models"
	models_da "annotater/internal/models/modelsDA"
	unit_test_utils "annotater/internal/tests/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm" // Ensure to import the testing package
)

type AnnotationRepositorySuite struct {
	suite.Suite
}

func (s *AnnotationRepositorySuite) TestAddAnottation(t provider.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := repo_adapter.NewAnotattionRepositoryAdapter(gormDB)
	validMarkup := unit_test_utils.VALID_MARKUP
	validMarkupDa, _ := models_da.ToDaMarkup(*validMarkup)
	tests := []struct {
		name        string
		setupMock   func()
		args        *models.Markup
		wantErr     bool
		expectedErr error
	}{

		{
			name: "Add annotation successfully",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "markups" ("page_data","class_label","creator_id","error_bb") VALUES ($1,$2,$3,$4) RETURNING "error_bb","id"`).
					WithArgs(validMarkupDa.PageData, validMarkupDa.ClassLabel, validMarkupDa.CreatorID, validMarkupDa.ErrorBB).
					WillReturnRows(
						sqlmock.NewRows([]string{"error_bb", "id"}).AddRow(validMarkupDa.ErrorBB, validMarkup.ID))

				mock.ExpectCommit()
			},
			args:    validMarkup,
			wantErr: false,
		},
		{
			name: "Add annotation with foreign key violation",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "markups" ("page_data","class_label","creator_id","error_bb") VALUES ($1,$2,$3,$4) RETURNING "error_bb","id"`).
					WithArgs(validMarkupDa.PageData, validMarkupDa.ClassLabel, validMarkupDa.CreatorID, validMarkupDa.ErrorBB).
					WillReturnError(gorm.ErrForeignKeyViolated)
				mock.ExpectRollback()
			},
			args:        validMarkup,
			wantErr:     true,
			expectedErr: models.ErrViolatingKeyAnnot,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.setupMock()
			err := repo.AddAnottation(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func (s *AnnotationRepositorySuite) TestDeleteAnotattion(t provider.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := repo_adapter.NewAnotattionRepositoryAdapter(gormDB)

	tests := []struct {
		name        string
		setupMock   func()
		args        uint64
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Delete annotation successfully",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "markups" WHERE id = $1`).
					WithArgs(unit_test_utils.TEST_BASIC_ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			args:    unit_test_utils.TEST_BASIC_ID,
			wantErr: false,
		},
		{
			name: "Delete annotation with error",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "markups" WHERE id = $1`).
					WithArgs(unit_test_utils.TEST_BASIC_ID).
					WillReturnError(errors.New("delete error"))
				mock.ExpectRollback()
			},
			args:    unit_test_utils.TEST_BASIC_ID,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.setupMock()
			err := repo.DeleteAnotattion(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func (s *AnnotationRepositorySuite) TestGetAnottationByID(t provider.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := repo_adapter.NewAnotattionRepositoryAdapter(gormDB)
	validMarkup := unit_test_utils.VALID_MARKUP
	validMarkupDa, _ := models_da.ToDaMarkup(*validMarkup)
	tests := []struct {
		name        string
		setupMock   func()
		args        uint64
		want        *models.Markup
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Get annotation by ID successfully",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "page_data", "error_bb", "class_label", "creator_id"}).
					AddRow(validMarkupDa.ID, validMarkupDa.PageData, validMarkupDa.ErrorBB, validMarkupDa.ClassLabel, validMarkupDa.CreatorID)
				mock.ExpectQuery(`SELECT * FROM "markups" WHERE id = $1 ORDER BY "markups"."id" LIMIT $2`).
					WithArgs(unit_test_utils.TEST_BASIC_ID, 1).
					WillReturnRows(rows)
			},
			args:    unit_test_utils.TEST_BASIC_ID,
			want:    validMarkup,
			wantErr: false,
		},
		{
			name: "Get annotation by ID with not found error",
			setupMock: func() {
				mock.ExpectQuery(`SELECT * FROM "markups" WHERE id = $1 ORDER BY "markups"."id" LIMIT $2`).
					WithArgs(unit_test_utils.TEST_BASIC_ID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			args:        unit_test_utils.TEST_BASIC_ID,
			want:        nil,
			wantErr:     true,
			expectedErr: models.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.setupMock()
			result, err := repo.GetAnottationByID(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func (s *AnnotationRepositorySuite) TestGetAnottationsByUserID(t provider.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := repo_adapter.NewAnotattionRepositoryAdapter(gormDB)
	validMarkup := unit_test_utils.VALID_MARKUP
	validMarkupDa, _ := models_da.ToDaMarkup(*validMarkup)
	tests := []struct {
		name        string
		setupMock   func()
		args        uint64
		want        []models.Markup
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Get annotations by User ID successfully",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "page_data", "error_bb", "class_label", "creator_id"}).
					AddRow(validMarkupDa.ID, validMarkupDa.PageData, validMarkupDa.ErrorBB, validMarkupDa.ClassLabel, validMarkupDa.CreatorID)
				mock.ExpectQuery(`SELECT * FROM "markups" WHERE creator_id = $1`).
					WithArgs(validMarkup.CreatorID).
					WillReturnRows(rows)
			},
			args:    validMarkup.CreatorID,
			want:    []models.Markup{*validMarkup},
			wantErr: false,
		},
		{
			name: "Get annotations by User ID with error",
			setupMock: func() {
				mock.ExpectQuery(`SELECT * FROM "markups" WHERE creator_id = $1`).
					WithArgs(validMarkupDa.CreatorID).
					WillReturnError(errors.New("query error"))
			},
			args:    validMarkup.CreatorID,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.setupMock()
			result, err := repo.GetAnottationsByUserID(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func (s *AnnotationRepositorySuite) TestGetAllAnottations(t provider.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := repo_adapter.NewAnotattionRepositoryAdapter(gormDB)
	validMarkup := unit_test_utils.VALID_MARKUP
	validMarkupDa, _ := models_da.ToDaMarkup(*validMarkup)
	tests := []struct {
		name        string
		setupMock   func()
		want        []models.Markup
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Get all annotations successfully",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "page_data", "error_bb", "class_label", "creator_id"}).
					AddRow(validMarkupDa.ID, validMarkupDa.PageData, validMarkupDa.ErrorBB, validMarkup.ClassLabel, validMarkup.CreatorID)
				mock.ExpectQuery(`SELECT \* FROM "markups"`).
					WillReturnRows(rows)
			},
			want:    []models.Markup{*validMarkup},
			wantErr: false,
		},
		{
			name: "Get all annotations with error",
			setupMock: func() {
				mock.ExpectQuery(`SELECT \* FROM "markups"`).
					WillReturnError(errors.New("query error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.setupMock()
			result, err := repo.GetAllAnottations()
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
