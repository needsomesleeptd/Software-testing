//go:build unit

package service_test

import (
	service "annotater/internal/bl/auth"

	repository "annotater/internal/bl/userService/userRepo"
	repo_adapter "annotater/internal/bl/userService/userRepo/userRepoAdapter"
	mock_repository "annotater/internal/mocks/bl/userService/userRepo"
	mock_auth_utils "annotater/internal/mocks/pkg/authUtils"
	"annotater/internal/models"
	auth_utils "annotater/internal/pkg/authUtils"
	unit_test_utils "annotater/internal/tests/utils"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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

func (s *AuthServiceSuite) Test_AuthService_Auth_Classic(t provider.T) {
	type fields struct {
		userRepo       repository.IUserRepository
		passwordHasher *mock_auth_utils.MockIPasswordHasher // IT IS BECAUSE OF NON DETERMINISTIC BEHAVIOUR
		tokenizer      auth_utils.ITokenHandler
		sqlMock        sqlmock.Sqlmock
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
				mother := unit_test_utils.NewUserObjectMother()
				user := mother.DefaultUser()
				userHashed := mother.DefaultUserHashedPassWd()

				f.passwordHasher.EXPECT().GenerateHash(user.Password).Return(userHashed.Password, nil)
				f.sqlMock.ExpectBegin()
				f.sqlMock.ExpectQuery(`INSERT INTO "users" ("login","password","name","surname","role","group") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`).
					WithArgs(user.Login, userHashed.Password, user.Name, user.Surname, user.Role, user.Group).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				f.sqlMock.ExpectCommit()
			},
			args:    args{*unit_test_utils.NewUserObjectMother().DefaultUser()},
			want:    nil,
			wantErr: false,
			err:     nil,
		},
		{
			name: "No login",
			prepare: func(f *fields, a *args) {

			},
			args:    args{*unit_test_utils.NewUserObjectMother().UserWithoutLogin()},
			want:    nil,
			wantErr: true,
			err:     errors.Wrapf(service.ErrNoLogin, service.ERR_LOGIN_STRF, unit_test_utils.TEST_EMPTY_LOGIN),
		},
	}
	ctrl := gomock.NewController(t)
	for _, tt := range tests {
		t.Title("Auth")
		t.Tags("auth")
		t.WithNewStep(tt.name, func(t provider.StepCtx) {

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			userStorage := repo_adapter.NewUserRepositoryAdapter(gormDB)
			passwordHasher := mock_auth_utils.NewMockIPasswordHasher(ctrl)
			tokenizer := auth_utils.NewJWTTokenHandler()

			if tt.prepare != nil {
				tt.prepare(&fields{
					userRepo:       userStorage,
					passwordHasher: passwordHasher,
					tokenizer:      tokenizer,
					key:            unit_test_utils.TEST_HASH_KEY,
					sqlMock:        mock,
				},
					&tt.args)
			}

			authService := service.NewAuthService(unit_test_utils.MockLogger, userStorage, passwordHasher, tokenizer, unit_test_utils.TEST_HASH_KEY)
			err = authService.SignUp(&tt.args.Candidate)

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

func (s *AuthServiceSuite) Test_AuthService_SignIn_Classic(t provider.T) {
	type args struct {
		Candidate models.User
	}

	type fields struct {
		userRepo       repository.IUserRepository
		passwordHasher *mock_auth_utils.MockIPasswordHasher
		tokenizer      *mock_auth_utils.MockITokenHandler
		sqlMock        sqlmock.Sqlmock
		key            string
	}

	tests := []struct {
		name    string
		prepare func(f *fields) // Adjusted to only require fields
		args    args
		getWant func(f *fields) string
		want    string
		wantErr bool
		err     error
	}{
		{
			name: "Valid User",
			prepare: func(f *fields) {
				user := unit_test_utils.NewUserObjectMother().DefaultUser()
				hashedPasswdUser := unit_test_utils.NewUserObjectMother().DefaultUserHashedPassWd()

				rows := sqlmock.NewRows([]string{"id", "login", "password", "name", "surname", "role", "group"}).
					AddRow(hashedPasswdUser.ID, hashedPasswdUser.Login, hashedPasswdUser.Password, hashedPasswdUser.Name, hashedPasswdUser.Surname, hashedPasswdUser.Role, hashedPasswdUser.Group)

				f.sqlMock.ExpectQuery(`SELECT * FROM "users" WHERE login = $1 ORDER BY "users"."id" LIMIT $2`).
					WithArgs(user.Login, 1).
					WillReturnRows(rows)

				f.passwordHasher.EXPECT().ComparePasswordhash(user.Password, hashedPasswdUser.Password).Return(nil)
				f.tokenizer.EXPECT().GenerateToken(*hashedPasswdUser, unit_test_utils.TEST_HASH_KEY).Return(unit_test_utils.TEST_VALID_TOKEN, nil)

			},
			args:    args{Candidate: *unit_test_utils.NewUserObjectMother().DefaultUser()},
			wantErr: false,
			getWant: func(f *fields) string {

				return unit_test_utils.TEST_VALID_TOKEN
			},
		},
		{
			name:    "No login",
			prepare: nil,
			args:    args{Candidate: *unit_test_utils.NewUserObjectMother().UserWithoutLogin()},
			getWant: nil,
			wantErr: true,
			err:     errors.Wrapf(service.ErrNoLogin, service.ERR_LOGIN_STRF, unit_test_utils.TEST_EMPTY_LOGIN),
		},
	}

	t.Title("SignIn")
	t.Tags("auth")
	ctrl := gomock.NewController(t)

	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			repo := repo_adapter.NewUserRepositoryAdapter(gormDB)
			tokenHandler := mock_auth_utils.NewMockITokenHandler(ctrl)     //nondeterm
			passwordHasher := mock_auth_utils.NewMockIPasswordHasher(ctrl) // non deterministic

			// Initialize fields for current test case
			fields := fields{
				userRepo:       repo,
				tokenizer:      tokenHandler,
				passwordHasher: passwordHasher,
				sqlMock:        mock,
				key:            unit_test_utils.TEST_HASH_KEY,
			}

			// Call prepare function if it exists
			if tt.prepare != nil {
				tt.prepare(&fields) // Pass the fields to the prepare function
			}
			if tt.getWant != nil {
				tt.want = tt.getWant(&fields)
			}

			authService := service.NewAuthService(unit_test_utils.MockLogger, repo, passwordHasher, tokenHandler, unit_test_utils.TEST_HASH_KEY)

			tokenStr, err := authService.SignIn(&tt.args.Candidate) // Call the SignIn method

			t.WithNewParameters("Candidate", tt.args.Candidate, "err", err)

			if tt.wantErr {
				t.Assert().Error(err)
				t.Assert().Equal(tt.err.Error(), err.Error())
			} else {
				t.Assert().NoError(err)
				// wasnot working as intended 0_0
				//lastDotIndexWant := strings.LastIndex(tt.want, ".")
				//expectedTokenData := tt.want[1:lastDotIndexWant]

				//lastDotIndexActual := strings.LastIndex(tokenStr, ".")
				//actualTokenData := tokenStr[1:lastDotIndexActual]
				t.Assert().Equal(tt.want, tokenStr)
			}

			// Ensure all expectations were met
			require.NoError(t, mock.ExpectationsWereMet())
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
