package integration_tests

//TODO:: split tests by files

import (
	nn_adapter "annotater/internal/bl/NN/NNAdapter"
	service "annotater/internal/bl/documentService"
	document_data_repo_adapter "annotater/internal/bl/documentService/documentDataRepo/documentDataRepo"
	document_metadata_repo_adapter "annotater/internal/bl/documentService/documentMetaDataRepo/documentMetaDataRepoAdapter"
	filesystem "annotater/internal/bl/documentService/reportDataRepo/reportDataRepoAdapter/filesytem"
	report_creator "annotater/internal/bl/reportCreatorService/reportCreator"
	mock_nn_model_handler "annotater/internal/mocks/bl/NN/NNAdapter/NNmodelhandler"
	"annotater/internal/models"
	models_da "annotater/internal/models/modelsDA"
	unit_test_utils "annotater/internal/tests/utils"
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ITRepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	pgContainer testcontainers.Container
	fs          filesystem.IFileSystem
}

func (suite *ITRepositoryTestSuite) SetupTest() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
			"POSTGRES_DB":       "test_db",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)
	suite.pgContainer = pgContainer

	// Get the host and port for the database
	host, err := pgContainer.Host(ctx)
	suite.Require().NoError(err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	suite.Require().NoError(err)

	// Open a new database connection for each test
	dsn := fmt.Sprintf("host=%s port=%s user=test_user password=test_password dbname=test_db sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	suite.Require().NoError(err)

	db.AutoMigrate(&models_da.Document{})

	suite.db = db
	suite.pgContainer = pgContainer
	//suite.fs = afero.NewOsFs()
}

func (suite *ITRepositoryTestSuite) TearDownTest() {

	suite.db.Migrator().DropTable(&models_da.Document{})
	ctx := context.Background()
	err := suite.pgContainer.Terminate(ctx)
	suite.Require().NoError(err)
}

type UsecaseRepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	pgContainer testcontainers.Container
}

func (suite *ITRepositoryTestSuite) TestUsecaseAddDocument() {
	var document *models.DocumentMetaData

	documentMetaRepo := document_metadata_repo_adapter.NewDocumentRepositoryAdapter(suite.db)
	documentRepo := document_data_repo_adapter.NewDocumentRepositoryAdapter("", ".temp", suite.fs)
	reportRepo := report_creator.NewPDFReportCreator()
	documentMetaMother := unit_test_utils.NewDocumentMetaDataBuilder()
	documentMother := unit_test_utils.NewMotherDocumentData().DefaultDocumentData()

	documentService := service.NewDocumentService()
	documentMeta := documentMetaMother.WithCreatorID(1).WithPageCount(1)
	insertedDocument := models.DocumentMetaData{}

	err := documentMetaRepo.AddDocument(documentMeta, document)
	suite.Require().NoError(err)

	document, err = documentMetaRepo.GetDocumentByID(id)
	suite.Require().NoError(err)
	suite.Assert().Equal(document.DocumentData, insertedDocument.DocumentData)
	suite.Assert().Equal(document.ID, id)
}

func (suite *ITRepositoryTestSuite) TestUsecaseLoadDocument() {
	var document *models.DocumentMetaData
	userRepo := document_repo_adapter.NewDocumentRepositoryAdapter(suite.db)

	handler := mock_nn_model_handler.NewMockIModelHandler(&gomock.Controller{})
	nn := nn_adapter.NewDetectionModel(handler)
	service := service.NewDocumentService(userRepo, nn)

	id := uuid.New() // Generate a new UUID
	insertedDocument := models.DocumentMetaData{ID: id, DocumentData: createPDFBuffer(TEST_VALID_PDF)}

	err := service.LoadDocument(insertedDocument)
	suite.Assert().NoError(err)

	document, err = userRepo.GetDocumentByID(id)
	suite.Require().NoError(err)
	suite.Assert().Equal(document.DocumentData, insertedDocument.DocumentData)
	suite.Assert().Equal(document.ID, id)
}

func (suite *ITRepositoryTestSuite) TestUsecaseDeleteDocumentID() {
	document := models.DocumentMetaData{}
	userRepo := document_repo_adapter.NewDocumentRepositoryAdapter(suite.db)

	id := uuid.New() // Generate a new UUID
	insertedDocument := models.DocumentMetaData{ID: id, DocumentData: createPDFBuffer(TEST_VALID_PDF)}

	err := userRepo.AddDocument(&insertedDocument)
	suite.Require().NoError(err)

	suite.Assert().NoError(suite.db.Table("documents").First(&document, models.DocumentMetaData{ID: id}).Error)
	err = userRepo.DeleteDocumentByID(id)
	suite.Require().NoError(err)

	suite.Assert().Error(suite.db.Table("documents").First(&document, models.DocumentMetaData{ID: id}).Error)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(ITRepositoryTestSuite))
}
