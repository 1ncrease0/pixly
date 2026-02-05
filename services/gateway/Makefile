.PHONY: swagger swagger-go
PATH := $(shell go env GOPATH)/bin:$(PATH)

swagger:
	swag init -g internal/infra/api/router.go -o docs --parseDependency --parseInternal

