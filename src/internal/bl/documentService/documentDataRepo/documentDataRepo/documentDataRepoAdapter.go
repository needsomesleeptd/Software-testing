package doc_data_repo_adapter

import (
	doc_data_repo "annotater/internal/bl/documentService/documentDataRepo"
	filesystem "annotater/internal/bl/documentService/reportDataRepo/reportDataRepoAdapter/filesytem"
	"annotater/internal/models"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type DocumentDataRepositoryAdapter struct {
	root          string
	fileExtension string //it is optional
	fs            filesystem.IFileSystem
}

func NewDocumentRepositoryAdapter(rootSrc string, ext string, fs filesystem.IFileSystem) doc_data_repo.IDocumentDataRepository {
	return &DocumentDataRepositoryAdapter{
		root:          rootSrc,
		fileExtension: ext,
		fs:            fs,
	}
}

func (repo *DocumentDataRepositoryAdapter) MakeDir(dir string) error {
	dirPath := fmt.Sprintf("%s/%s", repo.root, dir) + repo.fileExtension
	return repo.fs.MkdirAll(dirPath, os.FileMode(0755))
}

func (repo *DocumentDataRepositoryAdapter) Exists(path string) bool {
	fullPath := fmt.Sprintf("%s/%s", repo.root, path) + repo.fileExtension
	_, err := repo.fs.Stat(fullPath)

	return !repo.fs.IsNotExist(err)
}

func (repo *DocumentDataRepositoryAdapter) AddDocument(doc *models.DocumentData) error {
	if !repo.Exists(repo.root) {
		err := repo.MakeDir(repo.root)
		if err != nil {
			return errors.Wrap(err, "error in saving document data")
		}
	}

	filePath := fmt.Sprintf("%s/%s", repo.root, doc.ID) + repo.fileExtension

	err := repo.fs.WriteFile(filePath, doc.DocumentBytes, 0644)
	if err != nil {
		return errors.Wrap(err, "error in saving document data")
	}

	return nil
}

func (repo *DocumentDataRepositoryAdapter) DeleteDocumentByID(id uuid.UUID) error {
	filePath := fmt.Sprintf("%s/%s", repo.root, id) + repo.fileExtension
	err := repo.fs.Remove(filePath)
	if err != nil {
		return errors.Wrap(err, "error in deleting document data")
	}

	return nil
}

func (repo *DocumentDataRepositoryAdapter) GetDocumentByID(id uuid.UUID) (*models.DocumentData, error) {
	filePath := fmt.Sprintf("%s/%s", repo.root, id) + repo.fileExtension
	fileBytes, err := repo.fs.ReadFile(filePath)

	if err == os.ErrNotExist {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "error getting file")
	}

	documentData := &models.DocumentData{
		DocumentBytes: fileBytes,
		ID:            id,
	}
	return documentData, nil
}
