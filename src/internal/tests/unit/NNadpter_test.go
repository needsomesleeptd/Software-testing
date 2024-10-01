package service_test

import (
	nn_adapter "annotater/internal/bl/NN/NNAdapter"
	"annotater/internal/models"
	models_dto "annotater/internal/models/dto"
	unit_test_utils "annotater/internal/tests/utils"
	"errors"

	mock_nn_model_handler "annotater/internal/mocks/bl/NN/NNAdapter/NNmodelhandler"

	"github.com/golang/mock/gomock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/require"
)

type NNadapterSuite struct {
	suite.Suite
}

func (s *NNadapterSuite) TestDetectionModel_Predict(t provider.T) {
	type fields struct {
		modelHandler *mock_nn_model_handler.MockIModelHandler
	}
	type args struct {
		document models.DocumentData
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields)
		args    args
		want    []models.Markup
		wantErr bool
		err     error
	}{
		{
			name: "Valid Prediction",
			prepare: func(f *fields) {
				f.modelHandler.EXPECT().GetModelResp(gomock.Any()).Return([]models_dto.Markup{
					{ErrorBB: []float32{0.1, 0.2, 0.3}, ClassLabel: 1},
					{ErrorBB: []float32{0.3, 0.2, 0.1}, ClassLabel: 2},
				}, nil)
			},
			args: args{document: unit_test_utils.NewMotherDocumentData().DefaultDocumentData()},
			want: []models.Markup{
				{ErrorBB: []float32{0.1, 0.2, 0.3}, ClassLabel: 1},
				{ErrorBB: []float32{0.3, 0.2, 0.1}, ClassLabel: 2},
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "Error in Model Response",
			prepare: func(f *fields) {
				f.modelHandler.EXPECT().GetModelResp(gomock.Any()).Return(nil, errors.New("error in model response"))
			},
			args:    args{document: unit_test_utils.NewMotherDocumentData().DefaultDocumentData()},
			want:    nil,
			wantErr: true,
			err:     nn_adapter.ErrInModelPrediction,
		},
	}
	for _, tt := range tests {
		t.Title("Predict")
		t.Tags("NNadapter")
		//t.Parallel()
		ctrl := gomock.NewController(t)
		t.WithNewStep(tt.name, func(t provider.StepCtx) {

			defer ctrl.Finish()
			f := fields{
				modelHandler: mock_nn_model_handler.NewMockIModelHandler(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			s := nn_adapter.NewDetectionModel(f.modelHandler)
			markups, err := s.Predict(tt.args.document)
			if tt.wantErr {
				require.True(t, errors.Is(err, tt.err))
			} else {
				require.Nil(t, err)
				require.Equal(t, markups, tt.want)
			}

		})
	}
}
