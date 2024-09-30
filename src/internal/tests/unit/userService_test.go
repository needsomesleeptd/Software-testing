package service_test

import (
	service "annotater/internal/bl/userService"
	mock_repository "annotater/internal/mocks/bl/userService/userRepo"
	"annotater/internal/models"
	unit_test_utils "annotater/internal/tests/utils"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestUserService_ChangeUserRoleByLogin(t *testing.T) {
	type fields struct {
		userRepo *mock_repository.MockIUserRepository
	}
	type args struct {
		login string
		role  models.Role
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		prepare func(f *fields)
		errStr  error
	}{
		{
			name: "Changing role no err",
			prepare: func(f *fields) {
				f.userRepo.EXPECT().GetUserByLogin(unit_test_utils.TEST_VALID_LOGIN).Return(&models.User{Role: models.Admin}, nil)
				f.userRepo.EXPECT().UpdateUserByLogin(unit_test_utils.TEST_VALID_LOGIN, &models.User{Role: models.Controller}).Return(nil)
			},
			args:    args{login: unit_test_utils.TEST_VALID_LOGIN, role: models.Controller},
			wantErr: false,
			errStr:  nil,
		},
		{
			name: "Changing role getting err",
			prepare: func(f *fields) {
				f.userRepo.EXPECT().GetUserByLogin(unit_test_utils.TEST_VALID_LOGIN).Return(nil, unit_test_utils.ErrEmpty)
			},
			args:    args{login: unit_test_utils.TEST_VALID_LOGIN, role: models.Admin},
			wantErr: true,
			errStr:  errors.Wrap(unit_test_utils.ErrEmpty, fmt.Sprintf("error changing user role with login %v wanted role %v", unit_test_utils.TEST_VALID_LOGIN, models.Admin)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				userRepo: mock_repository.NewMockIUserRepository(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			s := service.NewUserService(unit_test_utils.MockLogger, f.userRepo)
			err := s.ChangeUserRoleByLogin(tt.args.login, tt.args.role)
			if tt.wantErr {
				require.Equal(t, tt.errStr.Error(), err.Error())
			} else {
				require.Nil(t, err)
			}

		})
	}
}

func TestUserService_GetAllUsers(t *testing.T) {
	type fields struct {
		userRepo *mock_repository.MockIUserRepository
	}
	type args struct {
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.User
		wantErr bool
		prepare func(f *fields)
		errStr  error
	}{
		{
			name: "Getting all users no err",
			prepare: func(f *fields) {
				userMother := unit_test_utils.NewUserObjectMother()
				user1 := userMother.NewUserWithLogin("log1")
				user2 := userMother.NewUserWithLogin("log2")
				f.userRepo.EXPECT().GetAllUsers().Return([]models.User{*user1, *user2}, nil)
			},
			want: []models.User{*unit_test_utils.NewUserObjectMother().NewUserWithLogin("log1"),
				*unit_test_utils.NewUserObjectMother().NewUserWithLogin("log2")},
			wantErr: false,
			errStr:  nil,
		},
		{
			name: "Getting all users with err",
			prepare: func(f *fields) {
				f.userRepo.EXPECT().GetAllUsers().Return(nil, unit_test_utils.ErrEmpty)
			},
			want:    nil,
			wantErr: true,
			errStr:  errors.Wrap(unit_test_utils.ErrEmpty, service.ERROR_GETTING_USERS_STR),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				userRepo: mock_repository.NewMockIUserRepository(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			s := service.NewUserService(unit_test_utils.MockLogger, f.userRepo)
			got, err := s.GetAllUsers()
			if tt.wantErr {
				require.Equal(t, tt.errStr.Error(), err.Error())
			} else {
				require.Nil(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
