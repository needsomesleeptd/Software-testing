package unit_test_utils

import (
	"annotater/internal/models"
	auth_utils "annotater/internal/pkg/authUtils"
)

const (
	TEST_HASH_KEY       = "test"
	TEST_VALID_LOGIN    = "login"
	TEST_VALID_PASSWORD = "passed"
	TEST_HASH_PASSWD    = "hashed_passwdf"
	TEST_VALID_TOKEN    = "token"

	TEST_EMPTY_LOGIN = ""

	TEST_EMPTY_PASSWD = ""
)

func forceGetHash(passwd string) string {
	hasher := auth_utils.NewPasswordHashCrypto()
	res, _ := hasher.GenerateHash(passwd)
	return res
}

type UserObjectMother struct{}

func NewUserObjectMother() *UserObjectMother {
	return &UserObjectMother{}
}

func (m *UserObjectMother) NewDefaultUser() *models.User {
	return &models.User{
		Login:    "default_login",
		Password: "default_password",
		Name:     "default name",
		Surname:  "default surname",
		Role:     models.Sender,
		Group:    "default group",
	}
}

func (m *UserObjectMother) NewUserWithLogin(login string) *models.User {
	return &models.User{
		Login:    login,
		Password: "default_password",
		Name:     "default name",
		Surname:  "default surname",
		Role:     models.Sender,
		Group:    "default group",
	}
}

func (m *UserObjectMother) NewUserWithPassword(password string) *models.User {
	return &models.User{
		Login:    "default_login",
		Password: password,
		Name:     "default name",
		Surname:  "default surname",
		Role:     models.Sender,
		Group:    "default group",
	}
}

func (m *UserObjectMother) NewUserWithLoginAndPassword(login string, password string) *models.User {
	return &models.User{
		Login:    login,
		Password: password,
		Name:     "default name",
		Surname:  "default surname",
		Role:     models.Sender,
		Group:    "default group",
	}
}

func (m *UserObjectMother) DefaultUser() *models.User {
	return &models.User{
		Login:    TEST_VALID_LOGIN,
		Password: TEST_VALID_PASSWORD,
		Name:     "default name",
		Surname:  "default surname",
		Role:     models.Sender,
		Group:    "default group",
	}
}

func (m *UserObjectMother) DefaultUserHashedPassWd() *models.User {
	return &models.User{
		Login:    TEST_VALID_LOGIN,
		Password: TEST_HASH_PASSWD,
		Name:     "default name",
		Surname:  "default surname",
		Role:     models.Sender,
		Group:    "default group",
	}
}

func (m *UserObjectMother) UserWithoutLogin() *models.User {
	return &models.User{
		Login:    TEST_EMPTY_LOGIN,
		Password: TEST_VALID_PASSWORD,
		Name:     "default name",
		Surname:  "default surname",
		Role:     models.Sender,
		Group:    "default group",
	}
}

func (m *UserObjectMother) UserWithoutPasswd() *models.User {
	return &models.User{
		Login:    TEST_VALID_LOGIN,
		Password: TEST_EMPTY_PASSWD,
		Name:     "default name",
		Surname:  "default surname",
		Role:     models.Sender,
		Group:    "default group",
	}
}
