package unit_test_utils

import (
	"annotater/internal/models"
	"bytes"

	"time"

	"github.com/google/uuid"
	"github.com/signintech/gopdf"
)

var TEST_VALID_PDF *gopdf.GoPdf = &gopdf.GoPdf{}

var TEST_VALID_UUID, _ = uuid.Parse("550e8400-e29b-41d4-a716-446655440000")

var TEST_DEFAULT_PAGE_COUNT = 12

var TEST_DEFAULT_CREATION_TIME = time.Date(2023, time.October, 1, 15, 0, 0, 0, time.UTC)

var TEST_DEFAULT_DOCUMENT_NAME = "default_doc"

func TEST_CreatePDFBuffer(pdf *gopdf.GoPdf) []byte {
	if pdf == nil {
		return []byte{1}
	}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	return buf.Bytes()
}

type DocumentMetaDataBuilder struct {
	documentMetaData *models.DocumentMetaData
}

func NewDocumentMetaDataBuilder() *DocumentMetaDataBuilder {
	return &DocumentMetaDataBuilder{documentMetaData: &models.DocumentMetaData{}}
}

func (b *DocumentMetaDataBuilder) WithDocumentID(documentID uuid.UUID) *DocumentMetaDataBuilder {
	b.documentMetaData.ID = documentID
	return b
}

func (b *DocumentMetaDataBuilder) WithPageCount(pageCount int) *DocumentMetaDataBuilder {
	b.documentMetaData.PageCount = pageCount
	return b
}

func (b *DocumentMetaDataBuilder) WithCreatorID(creatorID uint64) *DocumentMetaDataBuilder {
	b.documentMetaData.CreatorID = creatorID
	return b
}

func (b *DocumentMetaDataBuilder) Build() *models.DocumentMetaData {
	return b.documentMetaData
}

type MotherDocumentMeta struct{}

func NewMotherDocumentMeta() *MotherDocumentMeta {
	return &MotherDocumentMeta{}
}

func (m *MotherDocumentMeta) DefaultDocumentMeta() models.DocumentMetaData {
	return models.DocumentMetaData{
		ID:           TEST_VALID_UUID,
		CreatorID:    TEST_BASIC_ID,
		DocumentName: TEST_DEFAULT_DOCUMENT_NAME,
		PageCount:    TEST_DEFAULT_PAGE_COUNT,
		CreationTime: TEST_DEFAULT_CREATION_TIME,
	}
}

func (m *MotherDocumentMeta) DocumentMetaWithID(id uuid.UUID) models.DocumentMetaData {
	return models.DocumentMetaData{
		ID:           id,
		CreatorID:    TEST_BASIC_ID,
		DocumentName: TEST_DEFAULT_DOCUMENT_NAME,
		PageCount:    TEST_DEFAULT_PAGE_COUNT,
		CreationTime: TEST_DEFAULT_CREATION_TIME,
	}
}

func (m *MotherDocumentMeta) DocumentMetaWithCreatorID(creatorID uint64) models.DocumentMetaData {
	return models.DocumentMetaData{
		ID:           TEST_VALID_UUID,
		CreatorID:    creatorID,
		DocumentName: "doc_with_creator_id",
		PageCount:    TEST_DEFAULT_PAGE_COUNT,
		CreationTime: TEST_DEFAULT_CREATION_TIME,
	}
}

func (m *MotherDocumentMeta) DocumentMetaWithName(name string) models.DocumentMetaData {
	return models.DocumentMetaData{
		ID:           TEST_VALID_UUID,
		CreatorID:    TEST_BASIC_ID,
		DocumentName: name,
		PageCount:    TEST_DEFAULT_PAGE_COUNT,
		CreationTime: TEST_DEFAULT_CREATION_TIME,
	}
}

type MotherDocumentData struct{}

func NewMotherDocumentData() *MotherDocumentData {
	return &MotherDocumentData{}
}

func (m *MotherDocumentData) DefaultDocumentData() models.DocumentData {
	return models.DocumentData{
		ID:            TEST_VALID_UUID,
		DocumentBytes: TEST_CreatePDFBuffer(TEST_VALID_PDF),
	}
}

func (m *MotherDocumentData) InvalidDocumentData() models.DocumentData {
	return models.DocumentData{
		ID:            TEST_VALID_UUID,
		DocumentBytes: TEST_CreatePDFBuffer(nil),
	}
}
