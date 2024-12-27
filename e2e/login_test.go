package e2e_test

import (
	"errors"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/gavv/httpexpect/v2"
	"github.com/joho/godotenv"
	"github.com/xlzd/gotp"
)

var (
	expectLogin *httpexpect.Expect
)

// TestLogin initializes the test suite
func TestLogin(t *testing.T) {
	client := &http.Client{}

	expectLogin = httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:8111",
		Client:   client,
		Reporter: httpexpect.NewRequireReporter(t),
	})

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeLoginScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/login.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run login feature tests")
	}
}

// loginWith2FA defines the steps for login with 2FA
func loginWith2FA(ctx *godog.ScenarioContext) {
	var response *httpexpect.Response

	secret := os.Getenv("USER_SECRET")
	email := os.Getenv("USER_EMAIL")
	password := os.Getenv("USER_PASSWORD")

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

		response = expectLogin.Request(method, endpoint).
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
		response = expectLogin.Request(method, endpoint).
			WithJSON(map[string]string{
				"email": email,
				"code":  totp.Now(),
			}).Expect()

		return nil
	})

	ctx.Step(`^the response on /verify code should be (\d+)$`, func(statusCode int) error {
		if response == nil {
			return errors.New("response is nil")
		}
		response.Status(statusCode)
		return nil
	})

	ctx.Step(`^the response on /verify should match json:$`, func(expectedJSON *godog.DocString) error {
		if response == nil {
			return errors.New("response is nil")
		}
		response.JSON().Object().IsEqual(map[string]interface{}{
			"message": "Auth success!",
		})
		return nil
	})
}

// InitializeLoginScenario sets up the test environment
func InitializeLoginScenario(ctx *godog.ScenarioContext) {
	err := godotenv.Load("../app.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	loginWith2FA(ctx)
}
