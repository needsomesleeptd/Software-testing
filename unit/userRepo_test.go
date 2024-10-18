//go:build unit

package service_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	user_repo "annotater/internal/bl/userService/userRepo/userRepoAdapter"
	"annotater/internal/models"
	unit_test_utils "annotater/internal/tests/utils"
)

type UserRepoSuite struct {
	suite.Suite
}

func TestUserRepo(t *testing.T) {
	suite.RunSuite(t, new(UserRepoSuite))
}

func (s *UserRepoSuite) TestCreateUser(t provider.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := user_repo.NewUserRepositoryAdapter(gormDB)
	userMother := unit_test_utils.NewUserObjectMother()
	user := userMother.DefaultUser()
	tests := []struct {
		name        string
		setupMock   func()
		args        *models.User
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Create user successfully",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "users" ("login","password","name","surname","role","group") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`).
					WithArgs(user.Login, user.Password, user.Name, user.Surname, user.Role, user.Group).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			args:    user,
			wantErr: false,
		},
		{
			name: "Create user with error",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "users" ("login","password","name","surname","role","group") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`).
					WithArgs(user.Login, user.Password, user.Name, user.Surname, user.Role, user.Group).
					WillReturnError(errors.New("insert error"))
				mock.ExpectRollback()
			},
			args:    user,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Title(tt.name)
		t.Tags("userRepo")
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.setupMock()
			err := repo.CreateUser(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func (s *UserRepoSuite) TestGetUserByLogin(t provider.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := user_repo.NewUserRepositoryAdapter(gormDB)

	userMother := unit_test_utils.NewUserObjectMother()
	validUser := userMother.DefaultUser()

	tests := []struct {
		name        string
		setupMock   func()
		args        string
		want        *models.User
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Get user by ID successfully",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "login", "password", "name", "surname", "role", "group"}).
					AddRow(validUser.ID, validUser.Login, validUser.Password, validUser.Name, validUser.Surname, validUser.Role, validUser.Group)
				mock.ExpectQuery(`SELECT * FROM "users" WHERE login = $1 ORDER BY "users"."id" LIMIT $2`).
					WithArgs(validUser.Login, 1).
					WillReturnRows(rows)
			},
			args:    validUser.Login,
			want:    validUser,
			wantErr: false,
		},
		{
			name: "Get user by ID with error",
			setupMock: func() {
				mock.ExpectQuery(`SELECT * FROM "users" WHERE login = $1 ORDER BY "users"."id" LIMIT $2`).
					WithArgs(validUser.Login, 1).
					WillReturnError(errors.New("query error"))
			},
			args:    validUser.Login,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Title(tt.name)
		t.Tags("userRepo")
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.setupMock()
			user, err := repo.GetUserByLogin(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, user)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func (s *UserRepoSuite) TestUpdateUserByLogin(t provider.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := user_repo.NewUserRepositoryAdapter(gormDB)

	userMother := unit_test_utils.NewUserObjectMother()
	validUser := userMother.DefaultUser()

	tests := []struct {
		name        string
		setupMock   func()
		args        string
		user        *models.User
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Update user by login successfully",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users" SET "login"=$1,"password"=$2,"name"=$3,"surname"=$4,"group"=$5 WHERE login = $6`).
					WithArgs(validUser.Login, validUser.Password, validUser.Name, validUser.Surname, validUser.Group, validUser.Login).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			args:    validUser.Login,
			user:    validUser,
			wantErr: false,
		},
		{
			name: "Update user by login with error",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users" SET "login"=$1,"password"=$2,"name"=$3,"surname"=$4,"group"=$5 WHERE login = $6`).
					WithArgs(validUser.Login, validUser.Password, validUser.Name, validUser.Surname, validUser.Group, validUser.Login)
				mock.ExpectRollback()
			},
			args:    validUser.Login,
			user:    validUser,
			wantErr: true,
		},
	}
	t.Title("UpdateUserByLogin")
	t.Tags("userRepo")
	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.setupMock()
			err := repo.UpdateUserByLogin(tt.args, tt.user)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func (s *UserRepoSuite) TestGetAllUsers(t provider.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	repo := user_repo.NewUserRepositoryAdapter(gormDB)

	userMother := unit_test_utils.NewUserObjectMother()
	validUser1 := userMother.DefaultUser()
	validUser2 := userMother.DefaultUser()
	validUser2.ID = 2

	tests := []struct {
		name        string
		setupMock   func()
		want        []models.User
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Get all users successfully",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "login", "password", "name", "surname", "role", "group"}).
					AddRow(validUser1.ID, validUser1.Login, validUser1.Password, validUser1.Name, validUser1.Surname, validUser1.Role, validUser1.Group).
					AddRow(validUser2.ID, validUser2.Login, validUser2.Password, validUser2.Name, validUser2.Surname, validUser2.Role, validUser2.Group)
				mock.ExpectQuery(`SELECT * FROM "users"`).
					WillReturnRows(rows)
			},
			want:    []models.User{*validUser1, *validUser2},
			wantErr: false,
		},
		{
			name: "Get all users with error",
			setupMock: func() {
				mock.ExpectQuery(`SELECT * FROM "users"`).
					WillReturnError(errors.New("select error"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	t.Title("GetAllUsers")
	t.Tags("userRepo")
	for _, tt := range tests {
		t.WithNewStep(tt.name, func(t provider.StepCtx) {
			tt.setupMock()
			users, err := repo.GetAllUsers()
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, users)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
