# Запуск всех unit тестов проекта
test:
	rm -rf allure-results
	export ALLURE_OUTPUT_PATH="/home/andrew/uni/testing/Software-testing/src/internal/tests" && go test ./... --race --parallel 11; \
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

.PHONY: test allure report