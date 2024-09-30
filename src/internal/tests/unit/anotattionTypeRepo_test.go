package service_test

import (
	repo_adapter "annotater/internal/bl/anotattionTypeService/anottationTypeRepo/anotattionTypeRepoAdapter"
	"annotater/internal/models"
	unit_test_utils "annotater/internal/tests/utils"
	unit_test_mappers "annotater/internal/tests/utils/da"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	postgres2 "gorm.io/driver/postgres"
	"gorm.io/gorm" // Ensure to import the testing package
	"gorm.io/gorm/logger"
)

type AnnotationTypeRepositorySuite struct {
	suite.Suite
}

func (s *AnnotationTypeRepositorySuite) Test_AnnotationTypeRepository_GetAllAnnotationTypes(t provider.T) {
	type fields struct {
		db   *gorm.DB
		mock sqlmock.Sqlmock
	}
	type args struct{}

	tests := []struct {
		name       string
		fields     fields
		args       args
		beforeTest func(f *fields)
		want       []models.MarkupType
		wantErr    bool
		err        error
	}{
		{
			name: "Get all annotation types no err",
			beforeTest: func(f *fields) {
				markupTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
				markupType1 := markupTypeMother.NewMarkupTypeWithID(1)
				markupType2 := markupTypeMother.NewMarkupTypeWithID(2)
				f.mock.ExpectQuery("SELECT * FROM \"markup_types\"").WillReturnRows(
					unit_test_mappers.MapMarkupTypes(markupType1).
						AddRow(markupType2.ID, markupType2.Description, markupType2.CreatorID, markupType2.ClassName),
				)
			},
			want: []models.MarkupType{
				*unit_test_utils.NewMarkupTypeObjectMother().NewMarkupTypeWithID(1),
				*unit_test_utils.NewMarkupTypeObjectMother().NewMarkupTypeWithID(2),
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "Get all annotation types with err",
			beforeTest: func(f *fields) {
				f.mock.ExpectQuery("SELECT * FROM \"markup_types\"").WillReturnError(unit_test_utils.ErrEmpty)
			},
			want:    nil,
			wantErr: true,
			err:     errors.Wrapf(unit_test_utils.ErrEmpty, "Error in getting anotattion type db"),
		},
	}

	for _, tt := range tests {
		t.Title(tt.name)
		t.Tags("annotationTypeRepository")
		t.Run(tt.name, func(t provider.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open(postgres2.New(postgres2.Config{
				Conn: db,
			}), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			})

			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}

			// Adjusted to use gormDB instead of the mock
			fields := fields{
				db:   gormDB,
				mock: mock,
			}
			if tt.beforeTest != nil {
				tt.beforeTest(&fields)
			}

			repo := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)
			got, err := repo.GetAllAnottationTypes()

			if tt.wantErr {
				require.NotNil(t, err)
				require.Equal(t, tt.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func (s *AnnotationTypeRepositorySuite) Test_AnnotationTypeRepository_GetAnnotationTypeByID(t provider.T) {
	type fields struct {
		db   *gorm.DB
		mock sqlmock.Sqlmock
	}

	type args struct {
		id uint64
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		beforeTest func(f *fields)
		want       *models.MarkupType
		wantErr    bool
		err        error
	}{
		{
			name: "Get annotation type by id no err",
			beforeTest: func(f *fields) {
				markupTypeMother := unit_test_utils.NewMarkupTypeObjectMother()
				markupType := markupTypeMother.NewMarkupTypeWithID(unit_test_utils.TEST_BASIC_ID)
				f.mock.ExpectQuery("SELECT * FROM \"markup_types\" WHERE id = $1 AND \"markup_types\".\"id\" = $2 ORDER BY \"markup_types\".\"id\" LIMIT $3").
					WithArgs(unit_test_utils.TEST_BASIC_ID, unit_test_utils.TEST_BASIC_ID, 1).
					WillReturnRows(unit_test_mappers.MapMarkupTypes(markupType))
			},
			args:    args{id: unit_test_utils.TEST_BASIC_ID},
			want:    unit_test_utils.NewMarkupTypeObjectMother().NewMarkupTypeWithID(unit_test_utils.TEST_BASIC_ID),
			wantErr: false,
			err:     nil,
		},
		{
			name: "Get annotation type by id with err",
			beforeTest: func(f *fields) {
				f.mock.ExpectQuery("SELECT * FROM \"markup_types\" WHERE id = $1 AND \"markup_types\".\"id\" = $2 ORDER BY \"markup_types\".\"id\" LIMIT $3").
					WithArgs(unit_test_utils.TEST_BASIC_ID, unit_test_utils.TEST_BASIC_ID, 1).
					WillReturnError(unit_test_utils.ErrEmpty)
			},
			args:    args{id: unit_test_utils.TEST_BASIC_ID},
			want:    nil,
			wantErr: true,
			err:     errors.Wrap(unit_test_utils.ErrEmpty, "Error in getting anotattion type db"),
		},
	}

	for _, tt := range tests {
		t.Title(tt.name)
		t.Tags("annotationTypeRepository")
		t.Run(tt.name, func(t provider.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open(postgres2.New(postgres2.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
			}

			fields := fields{
				db:   gormDB,
				mock: mock,
			}

			if tt.beforeTest != nil {
				tt.beforeTest(&fields)
			}

			repo := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)
			got, err := repo.GetAnottationTypeByID(tt.args.id)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})

	}
}

func (s *AnnotationTypeRepositorySuite) Test_AnotattionRepositoryAdapter_GetAnottationsByUserID(t provider.T) {
	type fields struct {
		db   *gorm.DB
		mock sqlmock.Sqlmock
	}

	type args struct {
		id uint64
	}

	objectMother := unit_test_utils.NewMarkupTypeObjectMother()

	tests := []struct {
		name       string
		fields     fields
		args       args
		beforeTest func(f *fields)
		want       []models.MarkupType
		wantErr    bool
		err        error
	}{
		{
			name: "Get annotations by user ID no err",
			beforeTest: func(f *fields) {
				markUp1 := objectMother.NewMarkupTypeWithID(1)
				markUp2 := objectMother.NewMarkupTypeWithID(2)
				f.mock.ExpectQuery("SELECT * FROM \"markup_types\" WHERE creator_id = $1").
					WithArgs(unit_test_utils.TEST_BASIC_ID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "description", "creator_id", "class_name"}).
						AddRow(markUp1.ID, markUp1.Description, markUp1.CreatorID, markUp1.ClassName).
						AddRow(markUp2.ID, markUp2.Description, markUp2.CreatorID, markUp2.ClassName))
			},
			args: args{id: unit_test_utils.TEST_BASIC_ID},
			want: []models.MarkupType{
				*objectMother.NewMarkupTypeWithID(1),
				*objectMother.NewMarkupTypeWithID(2),
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "Get annotations by user ID with error from db",
			beforeTest: func(f *fields) {
				f.mock.ExpectQuery("SELECT * FROM \"markup_types\" WHERE creator_id = $1").
					WithArgs(unit_test_utils.TEST_BASIC_ID).
					WillReturnError(unit_test_utils.ErrEmpty)
			},
			args:    args{id: unit_test_utils.TEST_BASIC_ID},
			want:    nil,
			wantErr: true,
			err:     errors.Wrapf(unit_test_utils.ErrEmpty, "Error in getting anotattion type db"),
		},
	}

	for _, tt := range tests {
		t.Title(tt.name)
		t.Tags("anotationRepository")
		t.Run(tt.name, func(t provider.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open(postgres2.New(postgres2.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
			}

			fields := fields{
				db:   gormDB,
				mock: mock,
			}

			if tt.beforeTest != nil {
				tt.beforeTest(&fields)
			}

			repo := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)
			got, err := repo.GetAnottationTypesByUserID(tt.args.id)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func (s *AnnotationTypeRepositorySuite) Test_AnotattionTypeRepositoryAdapter_GetAnottationTypesByIDs(t provider.T) {

	type fields struct {
		db   *gorm.DB
		mock sqlmock.Sqlmock
	}

	objectMother := unit_test_utils.NewMarkupTypeObjectMother()

	tests := []struct {
		name       string
		fields     fields
		args       []uint64
		beforeTest func(f *fields)
		want       []models.MarkupType
		wantErr    bool
		err        error
	}{
		{
			name: "Get annotation types by IDs no error",
			beforeTest: func(f *fields) {
				markUp1 := objectMother.NewMarkupTypeWithID(1)
				markUp2 := objectMother.NewMarkupTypeWithID(2)

				f.mock.ExpectQuery(`SELECT * FROM "markup_types" WHERE "markup_types"."id" IN ($1,$2)`).
					WithArgs(markUp1.ID, markUp2.ID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "description", "creator_id", "class_name"}).
						AddRow(markUp1.ID, markUp1.Description, markUp1.CreatorID, markUp1.ClassName).
						AddRow(markUp2.ID, markUp2.Description, markUp2.CreatorID, markUp2.ClassName))
			},
			args: []uint64{1, 2},
			want: []models.MarkupType{
				*objectMother.NewMarkupTypeWithID(1),
				*objectMother.NewMarkupTypeWithID(2),
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "Get annotation types by IDs with error from db",
			beforeTest: func(f *fields) {
				f.mock.ExpectQuery(`SELECT * FROM "markup_types" WHERE "markup_types"."id" IN ($1,$2)`).
					WithArgs(uint64(1), uint64(2)).
					WillReturnError(unit_test_utils.ErrEmpty)
			},
			args:    []uint64{1, 2},
			want:    nil,
			wantErr: true,
			err:     errors.Wrapf(unit_test_utils.ErrEmpty, "Error in getting anotattion type db"),
		},
	}

	for _, tt := range tests {
		t.Title(tt.name)
		t.Tags("anotationTypeRepository")
		t.Run(tt.name, func(t provider.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open(postgres2.New(postgres2.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
			}

			fields := fields{
				db:   gormDB,
				mock: mock,
			}

			if tt.beforeTest != nil {
				tt.beforeTest(&fields)
			}

			repo := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)
			got, err := repo.GetAnottationTypesByIDs(tt.args)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func (s *AnnotationTypeRepositorySuite) Test_AnotattionTypeRepositoryAdapter_DeleteAnotattionType(t provider.T) {

	type fields struct {
		db   *gorm.DB
		mock sqlmock.Sqlmock
	}

	tests := []struct {
		name       string
		fields     fields
		args       uint64
		beforeTest func(f *fields)
		wantErr    bool
		err        error
	}{
		{
			name: "Delete annotation type no error",
			beforeTest: func(f *fields) {
				f.mock.ExpectBegin()
				f.mock.ExpectExec(`DELETE FROM "markup_types" WHERE id = $1`).
					WithArgs(uint64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
				f.mock.ExpectCommit()
			},
			args:    1,
			wantErr: false,
			err:     nil,
		},
		{
			name: "Delete annotation type with error",
			beforeTest: func(f *fields) {
				f.mock.ExpectBegin()
				f.mock.ExpectExec(`DELETE FROM "markup_types" WHERE id = $1`).
					WithArgs(uint64(1)).
					WillReturnError(unit_test_utils.ErrEmpty)
				f.mock.ExpectRollback()
			},
			args:    1,
			wantErr: true,
			err:     errors.Wrap(unit_test_utils.ErrEmpty, "Error in deleting anotattion type db"),
		},
	}

	for _, tt := range tests {
		t.Title(tt.name)
		t.Tags("anotationTypeRepository")
		t.Run(tt.name, func(t provider.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open(postgres2.New(postgres2.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
			}

			fields := fields{
				db:   gormDB,
				mock: mock,
			}

			if tt.beforeTest != nil {
				tt.beforeTest(&fields)
			}

			repo := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)
			err = repo.DeleteAnotattionType(tt.args)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (s *AnnotationTypeRepositorySuite) Test_AddAnottationType(t provider.T) {
	type fields struct {
		db   *gorm.DB
		mock sqlmock.Sqlmock
	}

	objectMother := unit_test_utils.NewMarkupTypeObjectMother()

	tests := []struct {
		name       string
		fields     fields
		args       *models.MarkupType
		beforeTest func(f *fields)
		wantErr    bool
		err        error
	}{
		{
			name: "Add annotation type no error",
			beforeTest: func(f *fields) {
				markUpType := objectMother.NewDefaultMarkupType()
				f.mock.ExpectBegin()
				f.mock.ExpectQuery(`INSERT INTO "markup_types" ("description","creator_id","class_name","id") VALUES ($1,$2,$3,$4) RETURNING "id"`).
					WithArgs(markUpType.Description, markUpType.CreatorID, markUpType.ClassName, markUpType.ID).
					WillReturnRows(
						sqlmock.NewRows([]string{"id"}).AddRow(markUpType.ID))
				f.mock.ExpectCommit()
			},
			args:    objectMother.NewDefaultMarkupType(),
			wantErr: false,
			err:     nil,
		},
		{
			name: "Add annotation type with duplicate error",
			beforeTest: func(f *fields) {
				markUpType := objectMother.NewDefaultMarkupType()
				f.mock.ExpectBegin()
				f.mock.ExpectQuery(`INSERT INTO "markup_types" ("description","creator_id","class_name","id") VALUES ($1,$2,$3,$4) RETURNING "id"`).
					WithArgs(markUpType.Description, markUpType.CreatorID, markUpType.ClassName, markUpType.ID).
					WillReturnError(gorm.ErrDuplicatedKey)
				f.mock.ExpectRollback()
			},
			args:    objectMother.NewDefaultMarkupType(),
			wantErr: true,
			err:     models.ErrDuplicateMarkupType,
		},
	}

	for _, tt := range tests {
		t.Title(tt.name)
		t.Tags("annotationTypeRepository")
		t.Run(tt.name, func(t provider.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)

			defer db.Close()

			gormDB, err := gorm.Open(postgres2.New(postgres2.Config{
				Conn: db,
			}), &gorm.Config{})
			require.NoError(t, err)

			fields := fields{
				db:   gormDB,
				mock: mock,
			}

			if tt.beforeTest != nil {
				tt.beforeTest(&fields)
			}

			repo := repo_adapter.NewAnotattionTypeRepositoryAdapter(gormDB)
			err = repo.AddAnottationType(tt.args)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAnotattionTypeRepoSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(AnnotationTypeRepositorySuite))
}
