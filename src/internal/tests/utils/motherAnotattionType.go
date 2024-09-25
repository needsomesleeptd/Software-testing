package unit_test_utils

import (
	"annotater/internal/models"
)

// MarkupTypeObjectMother is an Object Mother for creating MarkupType instances
type MarkupTypeObjectMother struct{}

func NewMarkupTypeObjectMother() *MarkupTypeObjectMother {
	return &MarkupTypeObjectMother{}
}

// NewDefaultMarkupType returns a default MarkupType instance
func (mom *MarkupTypeObjectMother) NewDefaultMarkupType() *models.MarkupType {
	return &models.MarkupType{
		ID:          TEST_BASIC_ID,
		Description: "default description",
		CreatorID:   int(TEST_BASIC_ID),
		ClassName:   "default",
	}
}

// NewMarkupTypeWithName returns a MarkupType instance with a specified name
func (mom *MarkupTypeObjectMother) NewMarkupTypeWithID(ID uint64) *models.MarkupType {
	return &models.MarkupType{
		ID:          ID,
		Description: "default description",
		CreatorID:   int(TEST_BASIC_ID),
		ClassName:   "default",
	}
}
