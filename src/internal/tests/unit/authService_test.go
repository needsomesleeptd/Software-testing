package service_test

import (
	service "annotater/internal/bl/auth"
	mock_repository "annotater/internal/mocks/bl/userService/userRepo"
	mock_auth_utils "annotater/internal/mocks/pkg/authUtils"
	"annotater/internal/models"
	unit_test_utils "annotater/internal/tests/utils"
	"testing"

	"github.com/pkg/errors"

	"github.com/golang/mock/gomock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type AuthServiceSuite struct {
	suite.Suite
}

func (s *AuthServiceSuite) Test_AuthService_Auth(t provider.T) {
	type fields struct {
		userRepo       *mock_repository.MockIUserRepository
		passwordHasher *mock_auth_utils.MockIPasswordHasher
		tokenizer      *mock_auth_utils.MockITokenHandler
		key            string
	}
	type args struct {
		Candidate models.User
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields, a *args)
		args    args
		wantErr bool
		err     error
		want    *models.User
	}{
		{
			name: "Valid Created User",
			prepare: func(f *fields, a *args) {
				user := unit_test_utils.NewUserObjectMother().DefaultUser()
				hashedPasswdUser := unit_test_utils.NewUserObjectMother().DefaultUserHashedPassWd()
				f.userRepo.EXPECT().CreateUser(hashedPasswdUser).Return(nil)
				f.passwordHasher.EXPECT().GenerateHash(user.Password).Return(hashedPasswdUser.Password, nil)

			},
			args:    args{*unit_test_utils.NewUserObjectMother().DefaultUser()},
			want:    nil,
			wantErr: false,
			err:     nil,
		},
		{
			name:    "No login",
			args:    args{*unit_test_utils.NewUserObjectMother().UserWithoutLogin()},
			want:    nil,
			wantErr: true,
			err:     errors.Wrapf(service.ErrNoLogin, service.ERR_LOGIN_STRF, unit_test_utils.TEST_EMPTY_LOGIN),
		},
		{
			name:    "No Passwd",
			args:    args{*unit_test_utils.NewUserObjectMother().UserWithoutPasswd()},
			want:    nil,
			wantErr: true,
			err:     errors.Wrapf(service.ErrNoPasswd, service.ERR_LOGIN_STRF, unit_test_utils.TEST_VALID_LOGIN),
		},
		{
			name: "Hash error",
			prepare: func(f *fields, a *args) {
				user := unit_test_utils.NewUserObjectMother().DefaultUser()
				f.passwordHasher.EXPECT().GenerateHash(user.Password).Return("", unit_test_utils.ErrEmpty)

			},
			args:    args{*unit_test_utils.NewUserObjectMother().DefaultUser()},
			want:    nil,
			wantErr: true,
			err:     errors.Wrapf(errors.Wrap(unit_test_utils.ErrEmpty, service.ErrGeneratingHash.Error()), service.ERR_LOGIN_STRF, unit_test_utils.TEST_VALID_LOGIN), // oh my god
		},
		{
			name: "CreateUser error",
			prepare: func(f *fields, a *args) {
				user := unit_test_utils.NewUserObjectMother().DefaultUser()
				hashedPasswdUser := unit_test_utils.NewUserObjectMother().DefaultUserHashedPassWd()
				f.passwordHasher.EXPECT().GenerateHash(user.Password).Return(hashedPasswdUser.Password, nil)
				f.userRepo.EXPECT().CreateUser(hashedPasswdUser).Return(errors.New(""))
				a.Candidate = *user
			},
			want:    nil,
			wantErr: true,
			err:     errors.Wrapf(errors.Wrap(unit_test_utils.ErrEmpty, service.ErrCreatingUser.Error()), service.ERR_LOGIN_STRF, unit_test_utils.TEST_VALID_LOGIN),
		},
	}
	for _, tt := range tests {
		t.Title("Auth")
		t.Tags("auth")
		ctrl := gomock.NewController(t)
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			userMockStorage := mock_repository.NewMockIUserRepository(ctrl)
			passwordHasherMock := mock_auth_utils.NewMockIPasswordHasher(ctrl)
			tokenizerMock := mock_auth_utils.NewMockITokenHandler(ctrl)
			if tt.prepare != nil {
				tt.prepare(&fields{
					userRepo:       userMockStorage,
					passwordHasher: passwordHasherMock,
					tokenizer:      tokenizerMock,
					key:            unit_test_utils.TEST_HASH_KEY,
				},
					&tt.args)
			}

			authService := service.NewAuthService(unit_test_utils.MockLogger, userMockStorage, passwordHasherMock, tokenizerMock, unit_test_utils.TEST_HASH_KEY)
			err := authService.SignUp(&tt.args.Candidate)

			t.WithNewParameters("Candidate", tt.args.Candidate, "err", err)
			if tt.wantErr {
				t.Assert().Error(err)
				t.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				t.Assert().NoError(err)
			}
		})
	}
}

func (s *AuthServiceSuite) Test_AuthService_SignIn(t provider.T) {
	type fields struct {
		userRepo       *mock_repository.MockIUserRepository
		passwordHasher *mock_auth_utils.MockIPasswordHasher
		tokenizer      *mock_auth_utils.MockITokenHandler
		key            string
	}
	type args struct {
		Candidate models.User
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields)
		args    args
		wantErr bool
		err     error
		want    string
	}{
		{
			name: "Valid User",
			prepare: func(f *fields) {
				user := unit_test_utils.NewUserObjectMother().DefaultUser()
				hashedPasswdUser := unit_test_utils.NewUserObjectMother().DefaultUserHashedPassWd()
				f.userRepo.EXPECT().GetUserByLogin(user.Login).Return(hashedPasswdUser, nil)
				f.passwordHasher.EXPECT().ComparePasswordhash(user.Password, hashedPasswdUser.Password).Return(nil)
				f.tokenizer.EXPECT().GenerateToken(*hashedPasswdUser, unit_test_utils.TEST_HASH_KEY).Return(unit_test_utils.TEST_HASH_PASSWD, nil)
			},
			args:    args{*unit_test_utils.NewUserObjectMother().DefaultUser()},
			want:    unit_test_utils.TEST_HASH_PASSWD,
			wantErr: false,
			err:     nil,
		},
		{
			name:    "No login",
			args:    args{*unit_test_utils.NewUserObjectMother().UserWithoutLogin()},
			want:    "",
			wantErr: true,
			err:     errors.Wrapf(service.ErrNoLogin, service.ERR_LOGIN_STRF, unit_test_utils.TEST_EMPTY_LOGIN),
		},
		{
			name:    "No Passwd",
			args:    args{*unit_test_utils.NewUserObjectMother().UserWithoutPasswd()},
			want:    "",
			wantErr: true,
			err:     errors.Wrapf(service.ErrNoPasswd, service.ERR_LOGIN_STRF, unit_test_utils.TEST_VALID_LOGIN),
		},
		{
			name: "User not found",
			prepare: func(f *fields) {
				user := unit_test_utils.NewUserObjectMother().DefaultUser()
				f.userRepo.EXPECT().GetUserByLogin(user.Login).Return(nil, errors.New(""))
			},
			args:    args{*unit_test_utils.NewUserObjectMother().DefaultUser()},
			want:    "",
			wantErr: true,
			err:     errors.Wrapf(unit_test_utils.ErrEmpty, service.ERR_LOGIN_STRF+":%v", unit_test_utils.TEST_VALID_LOGIN, service.ErrWrongLogin),
		},
		{
			name: "Password mismatch",
			prepare: func(f *fields) {
				user := unit_test_utils.NewUserObjectMother().DefaultUser()
				hashedPasswdUser := unit_test_utils.NewUserObjectMother().DefaultUserHashedPassWd()
				f.userRepo.EXPECT().GetUserByLogin(user.Login).Return(hashedPasswdUser, nil)
				f.passwordHasher.EXPECT().ComparePasswordhash(user.Password, hashedPasswdUser.Password).Return(errors.New(""))
			},
			args:    args{*unit_test_utils.NewUserObjectMother().DefaultUser()},
			want:    "",
			wantErr: true,
			err:     errors.Wrapf(unit_test_utils.ErrEmpty, service.ERR_LOGIN_STRF+":%v", unit_test_utils.TEST_VALID_LOGIN, service.ErrWrongPassword),
		},
	}
	t.Title("SignIn")
	t.Tags("auth")
	ctrl := gomock.NewController(t)
	for _, tt := range tests {
		//t.Parallel()
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			userMockStorage := mock_repository.NewMockIUserRepository(ctrl)
			passwordHasherMock := mock_auth_utils.NewMockIPasswordHasher(ctrl)
			tokenizerMock := mock_auth_utils.NewMockITokenHandler(ctrl)
			if tt.prepare != nil {
				tt.prepare(&fields{
					userRepo:       userMockStorage,
					passwordHasher: passwordHasherMock,
					tokenizer:      tokenizerMock,
					key:            unit_test_utils.TEST_HASH_KEY,
				})
			}

			authService := service.NewAuthService(unit_test_utils.MockLogger, userMockStorage, passwordHasherMock, tokenizerMock, unit_test_utils.TEST_HASH_KEY)
			tokenStr, err := authService.SignIn(&tt.args.Candidate) // removed the address-of operator

			t.WithNewParameters("Candidate", tt.args.Candidate, "err", err)
			if tt.wantErr {
				t.Assert().Error(err)
				t.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				t.Assert().NoError(err)
				t.Assert().Equal(tt.want, tokenStr)
			}
		})
	}
}

func TestAuthServiceSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(AuthServiceSuite))
}
