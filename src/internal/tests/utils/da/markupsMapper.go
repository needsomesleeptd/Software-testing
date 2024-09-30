package unit_test_mappers

import (
	"annotater/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func MapMarkupTypes(markupType *models.MarkupType) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "description", "creator_id", "class_name"}).
		AddRow(markupType.ID, markupType.Description, markupType.CreatorID, markupType.ClassName)
}
