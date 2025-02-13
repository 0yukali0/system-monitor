include .env

.PHONY: image
image:
	@go mod tidy
	@docker build -t ${CONTAINER_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-scratch -f docker/Dockerfile.scratch .
	@docker push ${CONTAINER_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-scratch
	@docker build -t ${CONTAINER_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-cuda -f docker/Dockerfile.cuda .
	@docker push ${CONTAINER_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-cuda

.PHONY: fmt
fmt:
	@gofmt -w -s .
	@golangci-lint run ./...