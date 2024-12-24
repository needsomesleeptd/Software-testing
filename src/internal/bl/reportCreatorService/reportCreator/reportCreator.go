package report_creator

import (
	"annotater/internal/models"
	bboxes_utils "annotater/tech_ui/utils/bboxes"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"strconv"

	"github.com/google/uuid"
)

var (
	texFileFilename = "check.tex"
	pdfFileFilename = "check.pdf"
	imgFolderPath   = "images/"
	fileFormat      = ".png"
	texHeader       = "\\documentclass{article}\n\\usepackage{graphicx}\n\n\\begin{document}\n\\section{Error Report}\n"
	texTail         = "\\end{document}"
)

type IReportCreator interface {
	CreateReport(reportID uuid.UUID, markups []models.Markup, markupTypes []models.MarkupType) (*models.ErrorReport, error)
}

type PDFReportCreator struct {
	folderPath string
}

func NewPDFReportCreator(workFolderPath string) IReportCreator { // TODO:: think about this
	err := os.MkdirAll(workFolderPath, 0777)
	if err != nil {
		panic(err)
	}
	return &PDFReportCreator{
		folderPath: workFolderPath,
	}
}

func (cr *PDFReportCreator) addImageLatex(imgPath string) string {
	return "\\newpage\n\\noindent\\includegraphics[width=0.9\\textwidth, height=0.9\\textheight]{" + imgPath + "}\n\n"
}

func (cr *PDFReportCreator) saveImagesWithBBs(filePathSave string, markups []models.Markup) ([]string, error) {
	imgPaths := make([]string, len(markups))
	for i, markup := range markups {
		img, _, err := image.Decode(bytes.NewReader(markup.PageData))
		if err != nil {
			return nil, err
		}
		boundingBoxImg := image.NewRGBA(img.Bounds())
		draw.Draw(boundingBoxImg, img.Bounds(), img, image.Point{}, draw.Src)
		boundingBoxColor := color.RGBA{255, 0, 0, 255}
		x1, y1, x2, y2 := int(markup.ErrorBB[0]), int(markup.ErrorBB[1]), int(markup.ErrorBB[2]), int(markup.ErrorBB[3])
		boundingBoxes := []bboxes_utils.BoundingBox{
			{
				XMin: x1,
				YMin: y1,
				XMax: x2,
				YMax: y2,
			},
		}
		bboxes_utils.DrawBoundingBoxes(boundingBoxImg, boundingBoxes, boundingBoxColor)

		imgFilePath := filePathSave + strconv.Itoa(i) + fileFormat

		outputFile, err := os.Create(imgFilePath)
		if err != nil {
			return nil, err
		}

		imgPaths[i] = imgFilePath
		defer outputFile.Close()
		err = png.Encode(outputFile, boundingBoxImg)
		if err != nil {
			return nil, err
		}
	}
	return imgPaths, nil
}

func (cr *PDFReportCreator) CreateReport(reportID uuid.UUID, markups []models.Markup, markupTypes []models.MarkupType) (*models.ErrorReport, error) {
	senderFolderPath := cr.createSenderFolder(reportID)
	if senderFolderPath == "" {
		return nil, fmt.Errorf("failed to create folder for report ID: %v", reportID)
	}

	hashMarkUpType := cr.createMarkupTypeMap(markupTypes)

	texFilePath, err := cr.createTexFile(senderFolderPath)
	if err != nil {
		return nil, err
	}

	imgPaths, err := cr.createImageFolder(senderFolderPath, markups)
	if err != nil {
		return nil, err
	}

	content, err := cr.generateLatexContent(markups, hashMarkUpType, imgPaths)
	if err != nil {
		return nil, err
	}

	if err := cr.writeTexFile(texFilePath, content); err != nil {
		return nil, err
	}

	if err := cr.compileLatex(texFilePath, senderFolderPath); err != nil {
		return nil, err
	}

	return cr.readPDF(senderFolderPath, reportID)
}

func (cr *PDFReportCreator) createSenderFolder(reportID uuid.UUID) string {
	path := cr.folderPath + "/" + reportID.String() + "/"
	if err := os.Mkdir(path, 0777); err != nil {
		return ""
	}
	return path
}

func (cr *PDFReportCreator) createMarkupTypeMap(markupTypes []models.MarkupType) map[uint64]models.MarkupType {
	hashMap := make(map[uint64]models.MarkupType)
	for _, markupType := range markupTypes {
		hashMap[markupType.ID] = markupType
	}
	return hashMap
}

func (cr *PDFReportCreator) createTexFile(senderFolderPath string) (string, error) {
	texFilePath := senderFolderPath + texFileFilename
	file, err := os.Create(texFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return texFilePath, nil
}

func (cr *PDFReportCreator) createImageFolder(senderFolderPath string, markups []models.Markup) ([]string, error) {
	imgFolderPath := senderFolderPath + imgFolderPath
	if err := os.MkdirAll(imgFolderPath, 0777); err != nil {
		return nil, err
	}
	return cr.saveImagesWithBBs(imgFolderPath, markups)
}

func (cr *PDFReportCreator) generateLatexContent(markups []models.Markup, hashMarkUpType map[uint64]models.MarkupType, imgPaths []string) (string, error) {
	content := texHeader
	for i := 0; i < len(markups); i++ {
		imgLatex := cr.addImageLatex(imgPaths[i])
		description := cr.getDescription(markups[i], hashMarkUpType)
		content += imgLatex + description + "\n"
	}
	return content + texTail, nil
}

func (cr *PDFReportCreator) getDescription(markup models.Markup, hashMarkUpType map[uint64]models.MarkupType) string {
	if markupType, exists := hashMarkUpType[markup.ClassLabel]; exists {
		return markupType.Description
	}
	return fmt.Sprintf("error not found description for label: %v", markup.ClassLabel)
}

func (cr *PDFReportCreator) writeTexFile(texFilePath, content string) error {
	file, err := os.Create(texFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func (cr *PDFReportCreator) compileLatex(texFilePath, senderFolderPath string) error {
	outputDirKey := fmt.Sprintf("-output-directory=%s", senderFolderPath)
	cmd := exec.Command("pdflatex", outputDirKey, texFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running latex compile: %v", err)
	}
	return nil
}

func (cr *PDFReportCreator) readPDF(senderFolderPath string, reportID uuid.UUID) (*models.ErrorReport, error) {
	pdfFilePath := senderFolderPath + pdfFileFilename
	pdfBytes, err := os.ReadFile(pdfFilePath)
	if err != nil {
		return nil, err
	}
	return &models.ErrorReport{
		DocumentID: reportID,
		ReportData: pdfBytes,
	}, nil
}
