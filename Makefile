include .env

.PHONY: image
image:
	@go mod tidy
	@docker build -t ${CONTAINER_REPO}/${IMAGE_NAME}:${IMAGE_TAG} .
	@docker push ${CONTAINER_REPO}/${IMAGE_NAME}:${IMAGE_TAG}

.PHONY: fmt
fmt:
	@gofmt -w -s .
	@golangci-lint run ./...