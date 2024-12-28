local-run:
	go run cmd/main.go

docker-run:
	docker compose up -d

docker-rm:
	docker compose down
	docker image rm software-testing-auth-app:latest alpine:latest

run-e2e: docker-run
	go test -v e2e/login_test.go && go test -v e2e/reset_test.go

.PHONY: local-run docker-run docker-rm run-e2e