package unit_test_utils

import (
	"annotater/internal/models"
	"bytes"
	"image"
	"image/png"
)

const (
	TEST_BASIC_ID uint64 = 20
)

func createPNGBuffer(img *image.RGBA) []byte {
	if img == nil {
		return nil
	}
	pngBuf := new(bytes.Buffer)
	png.Encode(pngBuf, img)
	return pngBuf.Bytes()
}

var TEST_VALID_PNG_IMG *image.RGBA = image.NewRGBA(image.Rect(0, 0, 100, 100))
var VALID_PNG_BUFFER = createPNGBuffer(TEST_VALID_PNG_IMG)
var INVALD_PNG_BUFFER = createPNGBuffer(nil)

var INVALID_BBS_PARAMS = []float32{-1.0, 1.0, 0.0, 1.0}
var VALID_BBS_PARAMS = []float32{1.0, 1.0, 0.0, 1.0}

var VALID_MARKUP = NewMarkupBuilder().
	WithErrorBB(VALID_BBS_PARAMS).
	WithClassLabel(1).
	WithPageData(VALID_PNG_BUFFER).Build()

type MarkupBuilder struct {
	markup *models.Markup
}

func NewMarkupBuilder() *MarkupBuilder {
	return &MarkupBuilder{markup: &models.Markup{}}
}

func (b *MarkupBuilder) WithErrorBB(errorBB []float32) *MarkupBuilder {
	b.markup.ErrorBB = errorBB
	return b
}

func (b *MarkupBuilder) WithPageData(pageData []byte) *MarkupBuilder {
	b.markup.PageData = pageData
	return b
}

func (b *MarkupBuilder) WithClassLabel(classLabel uint64) *MarkupBuilder {
	b.markup.ClassLabel = classLabel
	return b
}

func (b *MarkupBuilder) Build() *models.Markup {
	return b.markup
}
