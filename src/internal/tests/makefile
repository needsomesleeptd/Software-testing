# Запуск всех unit тестов проекта
test:
	rm -rf allure-results
	mkdir -p allure-results
	go test ./... --race --parallel 11; \
	if [ $$? -ne 0 ]; then \
		echo "Tests failed"; \
	fi

# Создание allure отчета по результатам теста
allure:
	mkdir -p allure-reports/history
	cp -R allure-reports/history allure-results
	rm -rf allure-reports
	allure generate allure-results -o allure-reports
	allure serve allure-results -p 4000

# Запуск всех unit тестов проекта с последующим созданием allure отчета
report: test allure


e2e:
	export CONFIG_PATH="/home/andrew/uni/testing/Software-testing/src/config/local.yaml" && \
	go test -tags=e2e ./end2end/... --race

ci-unit:
	export ALLURE_OUTPUT_PATH="../" && \
	export ALLURE_OUTPUT_FOLDER="./unit/allure-results" && \
	if ! go test -tags=unit ./unit/... --race; then \
		echo "export UNIT_FAILED=true" >> .env; \
		echo "Tests unit have failed, UNIT_FAILED variable written to .env."; \
	fi

ci-integration:
	export ALLURE_OUTPUT_PATH="../" && \
	export ALLURE_OUTPUT_FOLDER="./intergration_tests/allure-results" && \
	if  grep -q 'UNIT_FAILED' .env; then \
		export UNIT_FAILED="1"; \
		echo "UNIT_FAILED is set to 1."; \
		echo "export INTEGRATION_FAILED=true" >> .env; \
	fi && \
	if ! go test -tags=integration ./intergration_tests/... --race; then \
		echo "export INTEGRATION_FAILED=true" > .env; \
		echo "Tests integration have failed, INTEGRATION_FAILED variable written to .env."; \
	fi


ci-e2e:
	@export ALLURE_OUTPUT_PATH="../" && \
	export ALLURE_OUTPUT_FOLDER="./end2end/allure-results" && \
	export CONFIG_PATH="$(GITHUB_WORKSPACE)/src/config/local.yaml" && \
	if grep -q 'UNIT_FAILED' .env || grep -q 'INTEGRATION_FAILED' .env; then \
		export INTEGRATION_FAILED="1"; \
		echo "Set INTEGRATION_FAILED to 1."; \
	fi && \
	go test -tags=e2e ./end2end/... --race
	
ci-concat-reports:
	ls -la 
	ls -la ./intergration_tests
	mkdir allure-results
	cp unit/allure-results/* allure-results/
	cp intergration_tests/allure-results/* allure-results/
	cp end2end/allure-results/* allure-results/

.PHONY: test allure report ci-unit