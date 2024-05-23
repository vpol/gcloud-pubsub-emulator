BINARY=main
PROJECT_NAME=gcloud-pubsub-emulator
ENV ?= dev
LOG_LEVEL ?= debug
IMAGE_NAME ?= gcloud-pubsub-emulator
IMAGE_TAG ?= latest

GOLANGCI_VERSION ?= v1.58.0

all: clean lint build

TEST ?= ...

clean:
	@echo "--> Target directory clean up"
	rm -rf ./.build/target
	rm -f ${BINARY}

lint:
	@echo "--> Running linters"
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_VERSION} run -c .golangci.yml

build-docker:
	docker buildx build --tag "${IMAGE_NAME}:${IMAGE_TAG}" -f Dockerfile .

run-compose:
	docker-compose --project-name ${PROJECT_NAME} -f docker-compose.yml down
	docker-compose --project-name ${PROJECT_NAME} -f docker-compose.yml up --force-recreate --remove-orphans --renew-anon-volumes
