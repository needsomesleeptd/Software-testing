run:
	docker compose build 
	docker compose up -d

rerun-gatling:
	docker stop gatling-at-once gatling-per-second && docker rm gatling-at-once gatling-per-second
	docker compose up gatling-at-once gatling-per-second


docker-app:
	docker compose up gin-hello echo-hello prometheus grafana


rm:
	docker compose down
	docker image rm go_env:latest

gatling-trend:
	./gatling/scripts/trend.sh

gatling-delta:
	./gatling/scripts/delta.sh

gatling-stats: gatling-trend gatling-delta

.PHONY: s-bench run rerun-gatling rm gatling-trend gatling-delta gatling-stats