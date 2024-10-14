package service_test

import (
	rep_data_repo_adapter "annotater/internal/bl/documentService/reportDataRepo/reportDataRepoAdapter"
	mock_filesystem "annotater/internal/mocks/bl/documentService/reportDataRepo/reportDataRepoAdapter/filesytem"
	"annotater/internal/models"
	unit_test_utils "annotater/internal/tests/utils"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type DocumentRepoSuite struct {
	suite.Suite
}

func (s *DocumentRepoSuite) TestReportDataRepositoryAdapter(t provider.T) {
	// Set up the mock filesystem controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create the mocked filesystem
	mockFS := mock_filesystem.NewMockIFileSystem(ctrl)

	// Create the repository adapter
	repo := rep_data_repo_adapter.NewDocumentRepositoryAdapter(unit_test_utils.TEST_DEFAULT_ROOT, unit_test_utils.TEST_DEFAULT_EXT, mockFS)

	// Test suite for AddReport method

	tests := []struct {
		name        string
		report      models.ErrorReport
		expectedErr error
		prepareMock func()
		wantErr     bool
	}{
		{
			name:   "Success",
			report: unit_test_utils.NewErrReportMother().DefaultErrReport(),
			prepareMock: func() {
				report := unit_test_utils.NewErrReportMother().DefaultErrReport()
				fullPath := fmt.Sprintf("%s/%s", unit_test_utils.TEST_DEFAULT_ROOT, unit_test_utils.TEST_DEFAULT_ROOT) + unit_test_utils.TEST_DEFAULT_EXT
				filepath := fmt.Sprintf("%s/%s", unit_test_utils.TEST_DEFAULT_ROOT, report.DocumentID) + unit_test_utils.TEST_DEFAULT_EXT
				mockFS.EXPECT().Stat(fullPath).Return(nil, os.ErrNotExist)

				mockFS.EXPECT().MkdirAll(fullPath, os.FileMode(0755)).Return(nil) // Ensure MkdirAll is expected
				mockFS.EXPECT().WriteFile(filepath, report.ReportData, os.FileMode(0644)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "WriteFile Error",
			report:      unit_test_utils.NewErrReportMother().DefaultErrReport(),
			expectedErr: errors.Wrap(errors.New("write error"), "error in saving document data"),
			prepareMock: func() {
				mockFS.EXPECT().Stat(gomock.Any()).Return(nil, os.ErrNotExist)

				mockFS.EXPECT().MkdirAll(gomock.Any(), os.FileMode(0755)).Return(nil) // Ensure MkdirAll is expected
				mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), os.FileMode(0644)).Return(errors.New("write error"))
			},
			wantErr: true,
		},
	}
	t.Title("AddReport")
	t.Tag("reportRepo")
	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			// Prepare mocks for each test case
			tt.prepareMock()

			// Call the method
			err := repo.AddReport(&tt.report)

			// Assert the results
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func (s *DocumentRepoSuite) TestDeleteReportByID(t provider.T) {
	// Set up the mock filesystem controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create the mocked filesystem
	mockFS := mock_filesystem.NewMockIFileSystem(ctrl)

	// Create the repository adapter
	repo := rep_data_repo_adapter.NewDocumentRepositoryAdapter(unit_test_utils.TEST_DEFAULT_ROOT, unit_test_utils.TEST_DEFAULT_EXT, mockFS)

	tests := []struct {
		name        string
		id          uuid.UUID
		expectedErr error
		wantErr     bool
		prepareMock func()
	}{
		{
			name:    "Success",
			id:      unit_test_utils.TEST_VALID_UUID,
			wantErr: false,
			prepareMock: func() {
				mockFS.EXPECT().Remove(unit_test_utils.TEST_DEFAULT_ROOT + "/" + unit_test_utils.TEST_VALID_UUID.String() + unit_test_utils.TEST_DEFAULT_EXT).Return(nil)
			},
		},
		{
			name:        "Remove Error",
			id:          unit_test_utils.TEST_VALID_UUID,
			expectedErr: errors.Wrap(errors.New("remove error"), "error in deleting document data"),
			wantErr:     true,
			prepareMock: func() {
				mockFS.EXPECT().Remove(unit_test_utils.TEST_DEFAULT_ROOT + "/" + unit_test_utils.TEST_VALID_UUID.String() + unit_test_utils.TEST_DEFAULT_EXT).Return(errors.New("remove error"))
			},
		},
	}

	t.Title("DeleteReportByID")
	t.Tag("reportRepo")
	for _, tt := range tests {

		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.prepareMock()
			err := repo.DeleteReportByID(tt.id)

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *DocumentRepoSuite) TestGetDocumentByID(t provider.T) {
	// Set up the mock filesystem controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create the mocked filesystem
	mockFS := mock_filesystem.NewMockIFileSystem(ctrl)

	// Create the repository adapter
	repo := rep_data_repo_adapter.NewDocumentRepositoryAdapter("/fake/path", ".txt", mockFS)

	tests := []struct {
		name        string
		id          uuid.UUID
		reportData  []byte
		expectedErr error
		wantErr     bool
		prepareMock func()
	}{
		{
			name:       "Success",
			id:         unit_test_utils.TEST_VALID_UUID,
			reportData: []byte("Document data"),
			wantErr:    false,
			prepareMock: func() {
				mockFS.EXPECT().ReadFile("/fake/path/"+unit_test_utils.TEST_VALID_UUID.String()+".txt").
					Return([]byte("Document data"), nil)
			},
		},
		{
			name:        "File Not Found",
			id:          unit_test_utils.TEST_VALID_UUID,
			expectedErr: models.ErrNotFound,
			wantErr:     true,
			prepareMock: func() {
				mockFS.EXPECT().ReadFile("/fake/path/"+unit_test_utils.TEST_VALID_UUID.String()+".txt").
					Return(nil, os.ErrNotExist)
			},
		},
	}

	t.Title("GetDocumentByID")
	t.Tag("reportRepo")
	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.prepareMock()
			result, err := repo.GetDocumentByID(tt.id)

			if tt.wantErr {
				if tt.expectedErr == models.ErrNotFound {
					assert.Equal(t, models.ErrNotFound, err)
				} else {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &models.ErrorReport{
					DocumentID: tt.id,
					ReportData: tt.reportData,
				}, result)
			}
		})
	}
}

func TestReportRepoSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(DocumentRepoSuite))
}
