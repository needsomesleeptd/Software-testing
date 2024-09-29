package unit_test_utils

import (
	"annotater/internal/models"
)

type ErrReportMother struct{}

func NewErrReportMother() *ErrReportMother {
	return &ErrReportMother{}
}

func (e *ErrReportMother) DefaultErrReport() models.ErrorReport {
	return models.ErrorReport{
		DocumentID: TEST_VALID_UUID,
		ReportData: TEST_CreatePDFBuffer(TEST_VALID_PDF),
	}
}

func (e *ErrReportMother) InvalidErrReport() models.ErrorReport {
	return models.ErrorReport{
		DocumentID: TEST_VALID_UUID,
		ReportData: TEST_CreatePDFBuffer(nil),
	}
}
