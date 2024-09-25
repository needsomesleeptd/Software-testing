go-test-command:
	go test -v ./unit/...

test:
	rm -rf allure-results
	export ALLURE_OUTPUT_PATH="/home/andrew/uni/testing/Software-testing/src/internal/tests" && go test ./unit/... --race --parallel 11
	
allure:
	cp -R allure-reports/history allure-results
	rm -rf allure-reports
	allure generate allure-results -o allure-reports
	allure serve allure-results -p 4000

report: test allure

.PHONY: go-tests-template mockery-template go-test-command test allure report