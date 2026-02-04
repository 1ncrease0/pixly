.PHONY: gen gen-auth gen-art
PATH := $(shell go env GOPATH)/bin:$(PATH)

gen: gen-auth gen-art

gen-auth:
	protoc --go_out=. --go_opt=module=github.com/1ncrease0/pixly/proto \
		--go-grpc_out=. --go-grpc_opt=module=github.com/1ncrease0/pixly/proto \
		auth.proto

gen-art:
	protoc --go_out=. --go_opt=module=github.com/1ncrease0/pixly/proto \
		--go-grpc_out=. --go-grpc_opt=module=github.com/1ncrease0/pixly/proto \
		art.proto
