//go:build unit

package service_test

import (
	repo_adapter "annotater/internal/bl/documentService/documentMetaDataRepo/documentMetaDataRepoAdapter"
	"annotater/internal/models"
	unit_test_utils "annotater/internal/tests/utils"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DocumentMetaDataRepoSuite struct {
	suite.Suite
}

func (s *DocumentMetaDataRepoSuite) TestDocumentMetaDataRepositoryAdapter_AddDocument(t provider.T) {
	t.Title("AddDocument")

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := repo_adapter.NewDocumentRepositoryAdapter(gormDB)
	objectMother := unit_test_utils.NewMotherDocumentMeta()
	tests := []struct {
		name      string
		document  models.DocumentMetaData
		mockSetup func()
		wantErr   bool
	}{
		{
			name:     "Success",
			document: objectMother.DefaultDocumentMeta(),
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "documents" ("id","page_count","document_name","checks_count","creator_id","creation_time") VALUES ($1,$2,$3,$4,$5,$6)`).
					WithArgs(unit_test_utils.TEST_VALID_UUID, unit_test_utils.TEST_DEFAULT_PAGE_COUNT,
						unit_test_utils.TEST_DEFAULT_DOCUMENT_NAME, 0, unit_test_utils.TEST_BASIC_ID,
						unit_test_utils.TEST_DEFAULT_CREATION_TIME).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:     "Insertion error",
			document: objectMother.DefaultDocumentMeta(),
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "documents" ("id","page_count","document_name","checks_count","creator_id","creation_time") VALUES ($1,$2,$3,$4,$5,$6)`).
					WillReturnError(errors.New("insertion error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.mockSetup()

			err := repo.AddDocument(&tt.document)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *DocumentMetaDataRepoSuite) TestDocumentMetaDataRepositoryAdapter_GetDocumentByID(t provider.T) {
	t.Title("GetDocumentByID")

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := repo_adapter.NewDocumentRepositoryAdapter(gormDB)
	objectMother := unit_test_utils.NewMotherDocumentMeta()

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func()
		wantErr   bool
		want      models.DocumentMetaData
	}{
		{
			name: "Success",
			id:   unit_test_utils.TEST_VALID_UUID,
			mockSetup: func() {
				mock.ExpectQuery(`SELECT * FROM "documents" WHERE "documents"."id" = $1 ORDER BY "documents"."id" LIMIT $2`).
					WithArgs(unit_test_utils.TEST_VALID_UUID, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "page_count", "creator_id", "document_name", "creation_time"}).
						AddRow(unit_test_utils.TEST_VALID_UUID, unit_test_utils.TEST_DEFAULT_PAGE_COUNT,
							unit_test_utils.TEST_BASIC_ID, unit_test_utils.TEST_DEFAULT_DOCUMENT_NAME,
							unit_test_utils.TEST_DEFAULT_CREATION_TIME))
			},
			wantErr: false,
			want:    objectMother.DefaultDocumentMeta(),
		},
		{
			name: "Not Found",
			id:   unit_test_utils.TEST_VALID_UUID,
			mockSetup: func() {
				mock.ExpectQuery(`SELECT * FROM "documents" WHERE "documents"."id" = $1 ORDER BY "documents"."id" LIMIT $2`).
					WithArgs(unit_test_utils.TEST_VALID_UUID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
			want:    models.DocumentMetaData{},
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.mockSetup()

			result, err := repo.GetDocumentByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, *result)
			}
		})
	}
}

func (s *DocumentMetaDataRepoSuite) TestDocumentMetaDataRepositoryAdapter_DeleteDocumentByID(t provider.T) {
	t.Title("DeleteDocumentByID")
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := repo_adapter.NewDocumentRepositoryAdapter(gormDB)

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "Success",
			id:   unit_test_utils.TEST_VALID_UUID,
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "document_meta_data" WHERE "document_meta_data"."id" = $1`).
					WithArgs(unit_test_utils.TEST_VALID_UUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Deletion error",
			id:   uuid.New(),
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "document_meta_data" WHERE "document_meta_data"."id" = $1`).
					WithArgs(unit_test_utils.TEST_VALID_UUID).
					WillReturnError(errors.New("deletion error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.mockSetup()

			err := repo.DeleteDocumentByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *DocumentMetaDataRepoSuite) TestDocumentMetaDataRepositoryAdapter_GetDocumentsByCreatorID(t provider.T) {
	t.Title("GetDocumentsByCreatorID")
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := repo_adapter.NewDocumentRepositoryAdapter(gormDB)
	objectMother := unit_test_utils.NewMotherDocumentMeta()
	tests := []struct {
		name      string
		creatorID uint64
		mockSetup func()
		wantErr   bool
		want      []models.DocumentMetaData
	}{
		{
			name:      "Success",
			creatorID: 1,
			mockSetup: func() {
				mock.ExpectQuery(`SELECT * FROM "documents" WHERE creator_id = $1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "page_count", "creator_id", "document_name", "creation_time"}).
						AddRow(unit_test_utils.TEST_VALID_UUID, unit_test_utils.TEST_DEFAULT_PAGE_COUNT,
							unit_test_utils.TEST_BASIC_ID, unit_test_utils.TEST_DEFAULT_DOCUMENT_NAME,
							unit_test_utils.TEST_DEFAULT_CREATION_TIME).
						AddRow(unit_test_utils.TEST_VALID_UUID, unit_test_utils.TEST_DEFAULT_PAGE_COUNT,
							unit_test_utils.TEST_BASIC_ID, unit_test_utils.TEST_DEFAULT_DOCUMENT_NAME,
							unit_test_utils.TEST_DEFAULT_CREATION_TIME))
			},
			wantErr: false,
			want: []models.DocumentMetaData{
				objectMother.DefaultDocumentMeta(),
				objectMother.DefaultDocumentMeta(),
			},
		},
		{
			name:      "DB Error",
			creatorID: 1,
			mockSetup: func() {
				mock.ExpectQuery(`SELECT * FROM "documents" WHERE creator_id = $1`).
					WithArgs(1).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.mockSetup()

			result, err := repo.GetDocumentsByCreatorID(tt.creatorID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestDocumentMetaDataRepoRunner(t *testing.T) {
	suite.RunSuite(t, new(DocumentMetaDataRepoSuite))
}
