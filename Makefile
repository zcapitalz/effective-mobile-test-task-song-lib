.PHONY: run-compose

run-compose:
	docker compose -f ./deploy/docker-compose.yml up --build server

.PHONY: db-up

db-up:
	docker compose -f ./deploy/docker-compose.yml up db -d

.PHONY: db-destroy

db-destroy:
	docker compose -f ./deploy/docker-compose.yml stop db
	docker compose -f ./deploy/docker-compose.yml rm -fv db

.PHONY: generate-api-docs

generate-api-docs:
	swag init -generalInfo=app.go --dir=internal/app,internal/controllers/v1 --output=api/openapi-spec/v1 --parseInternal --parseDependency
	mv api/openapi-spec/v1/docs.go internal/controllers/v1

.PHONY: format-api-annotations

format-api-annotations:
	swag fmt internal/controllers