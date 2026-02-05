BINARY=main
PROJECT_NAME=gcloud-pubsub-emulator
ENV ?= dev
LOG_LEVEL ?= debug
IMAGE_NAME ?= gcloud-pubsub-emulator
IMAGE_TAG ?= latest

GOLANGCI_VERSION = `git -c 'versionsort.suffix=-' ls-remote --exit-code --refs --sort='version:refname' --tags https://github.com/golangci/golangci-lint 'v2.*.*' | tail --lines=1 | cut -d"/" -f3`

all: clean lint build

TEST ?= ...

clean:
	@echo "--> Target directory clean up"
	rm -rf ./.build/target
	rm -f ${BINARY}

lint:
	@echo "--> Running linters"
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_VERSION} run -c .golangci.yml

build-docker:
	docker buildx build --tag "${IMAGE_NAME}:${IMAGE_TAG}" -f Dockerfile .

run-compose:
	docker-compose --project-name ${PROJECT_NAME} -f docker-compose.yml down
	docker-compose --project-name ${PROJECT_NAME} -f docker-compose.yml up --force-recreate --remove-orphans --renew-anon-volumes
