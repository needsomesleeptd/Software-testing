package service_test

import (
	service "annotater/internal/bl/documentService"
	mock_repository_data "annotater/internal/mocks/bl/documentService/documentDataRepo"
	mock_repository_meta "annotater/internal/mocks/bl/documentService/documentMetaDataRepo"
	mock_report_data "annotater/internal/mocks/bl/documentService/reportDataRepo"
	mock_report_creator "annotater/internal/mocks/bl/reportCreatorService"
	unit_test_utils "annotater/internal/tests/utils"

	"annotater/internal/models"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type DocumentServiceSuite struct {
	suite.Suite
}

func (s *DocumentServiceSuite) TestDocumentService_LoadDocument(t provider.T) {
	type fields struct {
		repoMeta      *mock_repository_meta.MockIDocumentMetaDataRepository
		repoData      *mock_repository_data.MockIDocumentDataRepository
		reportData    *mock_report_data.MockIReportDataRepository
		reportCreator *mock_report_creator.MockIReportCreatorService
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
		t.Title(tt.name)
		t.Tags("document_service")
		//t.Parallel()
		t.Run(tt.name, func(t provider.T) {
			ctrl := gomock.NewController(t)
			docMetaRepoMock := mock_repository_meta.NewMockIDocumentMetaDataRepository(ctrl)
			docRepoMock := mock_repository_data.NewMockIDocumentDataRepository(ctrl)
			reportRepoMock := mock_report_data.NewMockIReportDataRepository(ctrl)
			reportServiceMock := mock_report_creator.NewMockIReportCreatorService(ctrl)
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

func (s *DocumentServiceSuite) Test_IDocumentService_GetDocumentsByCreatorID(t provider.T) {
	type fields struct {
		docMetaRepo   *mock_repository_meta.MockIDocumentMetaDataRepository
		docRepo       *mock_repository_data.MockIDocumentDataRepository
		reportRepo    *mock_report_data.MockIReportDataRepository
		reportService *mock_report_creator.MockIReportCreatorService
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
		t.Title(tt.name)
		t.Tags("document_service")
		//t.Parallel()
		t.Run(tt.name, func(t provider.T) {
			ctrl := gomock.NewController(t)
			docMetaRepoMock := mock_repository_meta.NewMockIDocumentMetaDataRepository(ctrl)
			docRepoMock := mock_repository_data.NewMockIDocumentDataRepository(ctrl)
			reportRepoMock := mock_report_data.NewMockIReportDataRepository(ctrl)
			reportServiceMock := mock_report_creator.NewMockIReportCreatorService(ctrl)
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

func TestDocumentServiceSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(DocumentServiceSuite))
}
