package e2e_test

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/gavv/httpexpect/v2"
	"github.com/xlzd/gotp"
)

var (
	expectReset *httpexpect.Expect
)

func TestResetPassword(t *testing.T) {
	client := &http.Client{}

	expectReset = httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:8111",
		Client:   client,
		Reporter: httpexpect.NewRequireReporter(nil),
	})

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeResetPasswordScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/reset.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run reset password feature tests")
	}
}

func resetPasswordWith2FA(ctx *godog.ScenarioContext) {
	var response *httpexpect.Response

	secret := os.Getenv("USER_SECRET")
	email := os.Getenv("USER_EMAIL")
	password := os.Getenv("USER_PASSWORD")
	newPassword := "new_password"

	if secret == "" {
		log.Fatal("USER_SECRET environment variable is not set")
	}
	if email == "" {
		log.Fatal("USER_EMAIL environment variable is not set")
	}
	if password == "" {
		log.Fatal("USER_PASSWORD environment variable is not set")
	}

	totp := gotp.NewDefaultTOTP(secret)

	ctx.Step(`^User send "([^"]*)" request to "([^"]*)"$`, func(method, endpoint string) error {

		response = expectReset.Request(method, endpoint).
			WithJSON(map[string]string{
				"email":    email,
				"password": password,
			}).
			Expect()

		return nil
	})

	ctx.Step(`^the response on /login code should be (\d+)$`, func(statusCode int) error {
		if response == nil {
			return errors.New("response is nil")
		}
		response.Status(statusCode)
		return nil
	})

	ctx.Step(`^the response on /login should match json:$`, func(expectedJSON *godog.DocString) error {
		if response == nil {
			return errors.New("response is nil")
		}
		response.JSON().Object().IsEqual(map[string]interface{}{
			"message": "Totp Qr saved on the disk, you must already have totp",
		})
		return nil
	})

	ctx.Step(`^user send "([^"]*)" request to "([^"]*)"$`, func(method, endpoint string) error {
		response = expectReset.Request(method, endpoint).
			WithJSON(map[string]string{
				"email":       email,
				"totpCode":    totp.Now(),
				"newPassword": newPassword,
				"oldPassword": password,
			}).Expect()
		return nil
	})

	ctx.Step(`^the response on /reset_password code should be (\d+)$`, func(statusCode int) error {
		response.Status(statusCode)
		return nil
	})

	ctx.Step(`^the response on /reset_password should match json:$`, func(expectedJSON *godog.DocString) error {
		if response == nil {
			return errors.New("response is nil")
		}
		fmt.Printf("response if %v", response.Body())
		response.JSON().Object().IsEqual(map[string]interface{}{
			"message": "Password changed successfully",
		})
		return nil
	})
}

func InitializeResetPasswordScenario(ctx *godog.ScenarioContext) {
	resetPasswordWith2FA(ctx)
}
