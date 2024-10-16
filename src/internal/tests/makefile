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


ci-unit:
	export ALLURE_OUTPUT_PATH="${GITHUB_WORKSPACE}" && \
 	export ALLURE_OUTPUT_FOLDER="./unit-allure" && \ 
	go test -tags=unit ./unit --race

ci-integration:
	export ALLURE_OUTPUT_PATH="${GITHUB_WORKSPACE}" && \
	export ALLURE_OUTPUT_FOLDER="./integration-allure" && \
	go test -tags=integration ./intergration_tests --race


ci-e2e:
	export ALLURE_OUTPUT_PATH="${GITHUB_WORKSPACE}" && \
	export ALLURE_OUTPUT_FOLDER="./e2e-allure" && \
	export CONFIG_PATH="../../config/local.yaml" && \
	go test -tags=e2e ./end2end --race
	
ci-concat-reports:
	mkdir allure-results
	cp unit-allure/* allure-results/
	cp integration-allure/* allure-results/
	cp e2e-allure/* allure-results/

.PHONY: test allure report