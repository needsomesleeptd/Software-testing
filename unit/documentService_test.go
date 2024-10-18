//go:build unit

package service_test

import (
	anot_type_repo "annotater/internal/bl/anotattionTypeService/anottationTypeRepo/anotattionTypeRepoAdapter"
	service "annotater/internal/bl/documentService"
	repository_data "annotater/internal/bl/documentService/documentDataRepo"
	repository_data_ad "annotater/internal/bl/documentService/documentDataRepo/documentDataRepo"
	repository_meta "annotater/internal/bl/documentService/documentMetaDataRepo"
	repository_meta_ad "annotater/internal/bl/documentService/documentMetaDataRepo/documentMetaDataRepoAdapter"
	report_data "annotater/internal/bl/documentService/reportDataRepo"
	report_data_ad "annotater/internal/bl/documentService/reportDataRepo/reportDataRepoAdapter"
	report_creator "annotater/internal/bl/reportCreatorService"
	mock_nn "annotater/internal/mocks/bl/NN"
	mock_repository_data "annotater/internal/mocks/bl/documentService/documentDataRepo"
	mock_repository_meta "annotater/internal/mocks/bl/documentService/documentMetaDataRepo"
	mock_report_data "annotater/internal/mocks/bl/documentService/reportDataRepo"
	mock_filesystem "annotater/internal/mocks/bl/documentService/reportDataRepo/reportDataRepoAdapter/filesytem"
	mock_rep_creator_service "annotater/internal/mocks/bl/reportCreatorService"
	mock_report_creator "annotater/internal/mocks/bl/reportCreatorService/reportCreator"
	unit_test_utils "annotater/internal/tests/utils"
	"fmt"
	"os"

	"annotater/internal/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DocumentServiceSuite struct {
	suite.Suite
}

func (s *DocumentServiceSuite) TestDocumentService_LoadDocument(t provider.T) {
	type fields struct {
		repoMeta      *mock_repository_meta.MockIDocumentMetaDataRepository
		repoData      *mock_repository_data.MockIDocumentDataRepository
		reportData    *mock_report_data.MockIReportDataRepository
		reportCreator *mock_rep_creator_service.MockIReportCreatorService
	}
	type args struct {
		documentMeta models.DocumentMetaData
		documentData models.DocumentData
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields)
		args    args
		wantErr bool
		errStr  error
	}{
		{
			name: "Successful Load Document",
			prepare: func(f *fields) {
				documentBuilderMeta := unit_test_utils.NewMotherDocumentMeta()
				documentBuilder := unit_test_utils.NewMotherDocumentData()
				documentMeta := documentBuilderMeta.DefaultDocumentMeta()
				document := documentBuilder.DefaultDocumentData()

				reportBuilder := unit_test_utils.NewErrReportMother()

				report := reportBuilder.DefaultErrReport()

				f.repoData.EXPECT().AddDocument(&document).Return(nil)
				f.repoMeta.EXPECT().AddDocument(&documentMeta).Return(nil)
				f.reportCreator.EXPECT().CreateReport(document).Return(&report, nil)
				f.reportData.EXPECT().AddReport(&report).Return(nil)
			},
			args: args{
				documentMeta: unit_test_utils.NewMotherDocumentMeta().DefaultDocumentMeta(),
				documentData: unit_test_utils.NewMotherDocumentData().DefaultDocumentData(),
			},
			wantErr: false,
			errStr:  nil,
		},
		{
			name: "invalid file format",
			prepare: func(f *fields) {

			},
			args: args{
				documentMeta: unit_test_utils.NewMotherDocumentMeta().DefaultDocumentMeta(),
				documentData: unit_test_utils.NewMotherDocumentData().InvalidDocumentData(),
			},
			wantErr: true,
			errStr:  errors.Wrapf(service.ErrDocumentFormat, "document with name %v", unit_test_utils.TEST_DEFAULT_DOCUMENT_NAME),
		},
	}
	for _, tt := range tests {
		t.Title("LoadDocument")
		t.Tags("document_service")
		ctrl := gomock.NewController(t)
		//t.Parallel()
		t.WithNewStep(tt.name, func(t provider.StepCtx) {

			docMetaRepoMock := mock_repository_meta.NewMockIDocumentMetaDataRepository(ctrl)
			docRepoMock := mock_repository_data.NewMockIDocumentDataRepository(ctrl)
			reportRepoMock := mock_report_data.NewMockIReportDataRepository(ctrl)
			reportServiceMock := mock_rep_creator_service.NewMockIReportCreatorService(ctrl)
			loggerMock := unit_test_utils.MockLogger

			if tt.prepare != nil {
				tt.prepare(&fields{
					repoMeta:      docMetaRepoMock,
					repoData:      docRepoMock,
					reportData:    reportRepoMock,
					reportCreator: reportServiceMock,
				})
			}

			documentService := service.NewDocumentService(loggerMock, docMetaRepoMock, docRepoMock, reportRepoMock, reportServiceMock)
			_, err := documentService.LoadDocument(tt.args.documentMeta, tt.args.documentData)
			if tt.wantErr {
				require.Equal(t, tt.errStr.Error(), err.Error())
			} else {
				require.Nil(t, err)
			}

		})
	}
}

// ДОБРО ПОЖАЛОВАТЬ В АД
func (s *DocumentServiceSuite) TestDocumentService_LoadDocument_Classic(t provider.T) {
	type fields struct {
		repoMeta      *repository_meta.IDocumentMetaDataRepository
		repoData      *repository_data.IDocumentDataRepository
		reportData    *report_data.IReportDataRepository
		reportCreator *mock_report_creator.MockIReportCreator
		sqlMock       sqlmock.Sqlmock
		fs            mock_filesystem.MockIFileSystem
		nn            *mock_nn.MockINeuralNetwork
	}
	type args struct {
		documentMeta models.DocumentMetaData
		documentData models.DocumentData
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields)
		args    args
		wantErr bool
		errStr  error
	}{
		{
			name: "Successful Load Document",
			prepare: func(f *fields) {
				documentBuilderMeta := unit_test_utils.NewMotherDocumentMeta()
				documentBuilder := unit_test_utils.NewMotherDocumentData()
				documentMeta := documentBuilderMeta.DefaultDocumentMeta()
				document := documentBuilder.DefaultDocumentData()

				reportBuilder := unit_test_utils.NewErrReportMother()
				report := reportBuilder.DefaultErrReport()

				markupBuilder := unit_test_utils.NewMarkupBuilder()
				markup := markupBuilder.Build()

				markupTypeBuilder := unit_test_utils.NewMarkupTypeObjectMother()
				markUpType := markupTypeBuilder.NewMarkupTypeWithID(markup.ID)
				// adding repo to the fs
				fullPath := fmt.Sprintf("%s/%s", unit_test_utils.TEST_DEFAULT_ROOT, unit_test_utils.TEST_DEFAULT_ROOT) + unit_test_utils.TEST_DEFAULT_EXT
				filepath := fmt.Sprintf("%s/%s", unit_test_utils.TEST_DEFAULT_ROOT, document.ID) + unit_test_utils.TEST_DEFAULT_EXT
				f.fs.EXPECT().Stat(fullPath).Return(nil, os.ErrNotExist)
				f.fs.EXPECT().MkdirAll(fullPath, os.FileMode(0755)).Return(nil) // Ensure MkdirAll is expected
				f.fs.EXPECT().WriteFile(filepath, document.DocumentBytes, os.FileMode(0644)).Return(nil)

				//adding repoMetaData
				f.sqlMock.ExpectBegin()
				f.sqlMock.ExpectExec(`INSERT INTO "documents" ("id","page_count","document_name","checks_count","creator_id","creation_time") VALUES ($1,$2,$3,$4,$5,$6)`).
					WithArgs(documentMeta.ID, documentMeta.PageCount,
						documentMeta.DocumentName, 0, documentMeta.CreatorID,
						documentMeta.CreationTime).
					WillReturnResult(sqlmock.NewResult(1, 1))
				f.sqlMock.ExpectCommit()
				//get markups
				f.nn.EXPECT().Predict(document).Return([]models.Markup{*markup}, nil)

				//get correspoding markupTypes
				f.sqlMock.ExpectQuery(`SELECT * FROM "markup_types" WHERE "markup_types"."id" = $1`).
					WithArgs(markup.ID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "description", "creator_id", "class_name"}).
						AddRow(markUpType.ID, markUpType.Description, markUpType.CreatorID, markUpType.ClassName))
				//create report
				f.reportCreator.EXPECT().CreateReport(document.ID, []models.Markup{*markup},
					[]models.MarkupType{*markUpType}).Return(&report, nil)

				// save the report
				f.fs.EXPECT().Stat(fullPath).Return(nil, os.ErrNotExist)
				f.fs.EXPECT().MkdirAll(fullPath, os.FileMode(0755)).Return(nil) // Ensure MkdirAll is expected
				f.fs.EXPECT().WriteFile(filepath, report.ReportData, os.FileMode(0644)).Return(nil)
			},
			args: args{
				documentMeta: unit_test_utils.NewMotherDocumentMeta().DefaultDocumentMeta(),
				documentData: unit_test_utils.NewMotherDocumentData().DefaultDocumentData(),
			},
			wantErr: false,
			errStr:  nil,
		},
		{
			name: "invalid file format",
			prepare: func(f *fields) {

			},
			args: args{
				documentMeta: unit_test_utils.NewMotherDocumentMeta().DefaultDocumentMeta(),
				documentData: unit_test_utils.NewMotherDocumentData().InvalidDocumentData(),
			},
			wantErr: true,
			errStr:  errors.Wrapf(service.ErrDocumentFormat, "document with name %v", unit_test_utils.TEST_DEFAULT_DOCUMENT_NAME),
		},
	}
	for _, tt := range tests {
		t.Title("LoadDocument")
		t.Tags("document_service")
		ctrl := gomock.NewController(t)
		//t.Parallel()
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			docMetaRepo := repository_meta_ad.NewDocumentRepositoryAdapter(gormDB)
			fsMock := mock_filesystem.NewMockIFileSystem(ctrl)
			docRepo := repository_data_ad.NewDocumentRepositoryAdapter(unit_test_utils.TEST_DEFAULT_ROOT, unit_test_utils.TEST_DEFAULT_EXT, fsMock)
			reportRepo := report_data_ad.NewDocumentRepositoryAdapter(unit_test_utils.TEST_DEFAULT_ROOT, unit_test_utils.TEST_DEFAULT_EXT, fsMock)

			reportWorker := mock_report_creator.NewMockIReportCreator(ctrl)
			nn := mock_nn.NewMockINeuralNetwork(ctrl)
			mockAnnotTypeRepo := anot_type_repo.NewAnotattionTypeRepositoryAdapter(gormDB)

			reportService := report_creator.NewDocumentService(unit_test_utils.MockLogger, nn, mockAnnotTypeRepo, reportWorker)

			if tt.prepare != nil {
				tt.prepare(&fields{
					repoMeta:      &docMetaRepo,
					repoData:      &docRepo,
					reportData:    &reportRepo,
					reportCreator: reportWorker,
					sqlMock:       mock,
					nn:            nn,
					fs:            *fsMock,
				})
			}
			documentService := service.NewDocumentService(unit_test_utils.MockLogger, docMetaRepo, docRepo, reportRepo, reportService)
			_, err = documentService.LoadDocument(tt.args.documentMeta, tt.args.documentData)
			if tt.wantErr {
				require.Equal(t, tt.errStr.Error(), err.Error())
			} else {
				require.Nil(t, err)
			}

		})
	}
}

func (s *DocumentServiceSuite) Test_IDocumentService_GetDocumentsByCreatorID(t provider.T) {
	type fields struct {
		docMetaRepo   *mock_repository_meta.MockIDocumentMetaDataRepository
		docRepo       *mock_repository_data.MockIDocumentDataRepository
		reportRepo    *mock_report_data.MockIReportDataRepository
		reportService *mock_rep_creator_service.MockIReportCreatorService
		logger        *logrus.Logger
	}
	type args struct {
		creatorID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields)
		args    args
		want    []models.DocumentMetaData
		wantErr bool
		err     error
	}{
		{
			name: "Valid creator ID",
			prepare: func(f *fields) {
				document1 := unit_test_utils.NewMotherDocumentMeta().DocumentMetaWithName("doc1")
				document2 := unit_test_utils.NewMotherDocumentMeta().DocumentMetaWithName("doc2")
				documents := []models.DocumentMetaData{document1, document2}
				f.docMetaRepo.EXPECT().GetDocumentsByCreatorID(unit_test_utils.TEST_BASIC_ID).Return(documents, nil)
			},
			args: args{creatorID: unit_test_utils.TEST_BASIC_ID},
			want: []models.DocumentMetaData{unit_test_utils.NewMotherDocumentMeta().DocumentMetaWithName("doc1"),
				unit_test_utils.NewMotherDocumentMeta().DocumentMetaWithName("doc2")},
			wantErr: false,
			err:     nil,
		},
		{
			name: "Invalid creator ID",
			prepare: func(f *fields) {
				f.docMetaRepo.EXPECT().GetDocumentsByCreatorID(unit_test_utils.TEST_BASIC_ID).Return(nil, errors.New("invalid creator ID"))
			},
			args:    args{creatorID: unit_test_utils.TEST_BASIC_ID},
			want:    nil,
			wantErr: true,
			err:     errors.Wrapf(errors.New("invalid creator ID"), service.DOCUMENT_GET_ERR_CREATOR_STRF, unit_test_utils.TEST_BASIC_ID),
		},
	}
	for _, tt := range tests {
		t.Title("GetDocumentsByCreatorID")
		t.Tags("document_service")
		//t.Parallel()
		ctrl := gomock.NewController(t)
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			docMetaRepoMock := mock_repository_meta.NewMockIDocumentMetaDataRepository(ctrl)
			docRepoMock := mock_repository_data.NewMockIDocumentDataRepository(ctrl)
			reportRepoMock := mock_report_data.NewMockIReportDataRepository(ctrl)
			reportServiceMock := mock_rep_creator_service.NewMockIReportCreatorService(ctrl)
			loggerMock := unit_test_utils.MockLogger

			if tt.prepare != nil {
				tt.prepare(&fields{
					docMetaRepo:   docMetaRepoMock,
					docRepo:       docRepoMock,
					reportRepo:    reportRepoMock,
					reportService: reportServiceMock,
					logger:        loggerMock,
				})
			}

			documentService := service.NewDocumentService(loggerMock, docMetaRepoMock, docRepoMock, reportRepoMock, reportServiceMock)
			documents, err := documentService.GetDocumentsByCreatorID(tt.args.creatorID)

			t.WithNewParameters("creatorID", tt.args.creatorID, "err", err)
			if tt.wantErr {
				t.Assert().Error(err)
				t.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				t.Assert().NoError(err)
				t.Assert().Equal(tt.want, documents)
			}
		})
	}
}

func (s *DocumentServiceSuite) TestGetDocumentByID(t provider.T) {
	type fields struct {
		docMetaRepo   *mock_repository_meta.MockIDocumentMetaDataRepository
		docRepo       *mock_repository_data.MockIDocumentDataRepository
		reportRepo    *mock_report_data.MockIReportDataRepository
		reportService *mock_rep_creator_service.MockIReportCreatorService
		logger        *logrus.Logger
	}
	tests := []struct {
		name       string
		documentID uuid.UUID
		setupMock  func(f *fields)
		want       models.DocumentData
		wantErrMsg string
		wantErr    bool
	}{
		{
			name:       "Success",
			documentID: unit_test_utils.TEST_VALID_UUID,
			setupMock: func(f *fields) {
				mockDoc := unit_test_utils.NewMotherDocumentData().DefaultDocumentData()
				f.docRepo.EXPECT().GetDocumentByID(unit_test_utils.TEST_VALID_UUID).Return(&mockDoc, nil)
			},
			want:    unit_test_utils.NewMotherDocumentData().DefaultDocumentData(),
			wantErr: false,
		},
		{
			name:       "Document Not Found",
			documentID: unit_test_utils.TEST_VALID_UUID,
			setupMock: func(f *fields) {
				f.docRepo.EXPECT().GetDocumentByID(unit_test_utils.TEST_VALID_UUID).Return(nil, models.ErrNotFound)
			},
			want:       models.DocumentData{},
			wantErr:    true,
			wantErrMsg: "",
		},
	}
	t.Title("GetDocumentByID")
	t.Tags("document_service")
	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		t.WithNewStep(tt.name, func(t provider.StepCtx) {

			docMetaRepoMock := mock_repository_meta.NewMockIDocumentMetaDataRepository(ctrl)
			docRepoMock := mock_repository_data.NewMockIDocumentDataRepository(ctrl)
			reportRepoMock := mock_report_data.NewMockIReportDataRepository(ctrl)
			reportServiceMock := mock_rep_creator_service.NewMockIReportCreatorService(ctrl)
			loggerMock := unit_test_utils.MockLogger

			if tt.setupMock != nil {
				tt.setupMock(&fields{
					docMetaRepo:   docMetaRepoMock,
					docRepo:       docRepoMock,
					reportRepo:    reportRepoMock,
					reportService: reportServiceMock,
					logger:        loggerMock,
				})
			}

			documentService := service.NewDocumentService(loggerMock, docMetaRepoMock, docRepoMock, reportRepoMock, reportServiceMock)
			result, err := documentService.GetDocumentByID(tt.documentID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tt.want, result)
			}
		})
	}
}

func (s *DocumentServiceSuite) TestGetReportByID(t provider.T) {
	type fields struct {
		docMetaRepo   *mock_repository_meta.MockIDocumentMetaDataRepository
		docRepo       *mock_repository_data.MockIDocumentDataRepository
		reportRepo    *mock_report_data.MockIReportDataRepository
		reportService *mock_rep_creator_service.MockIReportCreatorService
		logger        *logrus.Logger
	}
	tests := []struct {
		name       string
		documentID uuid.UUID
		setupMock  func(f *fields)
		want       models.ErrorReport
		wantErrMsg string
		wantErr    bool
	}{
		{
			name:       "Success",
			documentID: unit_test_utils.TEST_VALID_UUID,
			setupMock: func(f *fields) {
				mockDoc := unit_test_utils.NewErrReportMother().DefaultErrReport()
				f.reportRepo.EXPECT().GetDocumentByID(unit_test_utils.TEST_VALID_UUID).Return(&mockDoc, nil)
			},
			want:    unit_test_utils.NewErrReportMother().DefaultErrReport(),
			wantErr: false,
		},
		{
			name:       "Document Not Found",
			documentID: unit_test_utils.TEST_VALID_UUID,
			setupMock: func(f *fields) {
				f.reportRepo.EXPECT().GetDocumentByID(unit_test_utils.TEST_VALID_UUID).Return(nil, models.ErrNotFound)
			},
			want:       models.ErrorReport{},
			wantErr:    true,
			wantErrMsg: "",
		},
	}
	t.Title("GetDocumentByID")
	t.Tags("document_service")
	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		t.WithNewStep(tt.name, func(t provider.StepCtx) {

			docMetaRepoMock := mock_repository_meta.NewMockIDocumentMetaDataRepository(ctrl)
			docRepoMock := mock_repository_data.NewMockIDocumentDataRepository(ctrl)
			reportRepoMock := mock_report_data.NewMockIReportDataRepository(ctrl)
			reportServiceMock := mock_rep_creator_service.NewMockIReportCreatorService(ctrl)
			loggerMock := unit_test_utils.MockLogger

			if tt.setupMock != nil {
				tt.setupMock(&fields{
					docMetaRepo:   docMetaRepoMock,
					docRepo:       docRepoMock,
					reportRepo:    reportRepoMock,
					reportService: reportServiceMock,
					logger:        loggerMock,
				})
			}

			documentService := service.NewDocumentService(loggerMock, docMetaRepoMock, docRepoMock, reportRepoMock, reportServiceMock)
			result, err := documentService.GetReportByID(tt.documentID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tt.want, result)
			}
		})
	}
}

func TestDocumentServiceSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(DocumentServiceSuite))
}
