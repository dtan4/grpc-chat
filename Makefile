OS   := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

BIN_DIR            := $(shell pwd)/bin
PROTOC_GEN_GO      := $(abspath $(BIN_DIR)/protoc-gen-go)
PROTOC_GEN_GO_GRPC := $(abspath $(BIN_DIR)/protoc-gen-go-grpc)

PROTOC_GEN_GO_VERSION      = 1.27.1
PROTOC_GEN_GO_GRPC_VERSION = 1.1.0

$(PROTOC_GEN_GO):
	curl -sSL https://github.com/protocolbuffers/protobuf-go/releases/download/v$(PROTOC_GEN_GO_VERSION)/protoc-gen-go.v$(PROTOC_GEN_GO_VERSION).$(OS).$(ARCH).tar.gz | tar -C $(BIN_DIR) -xzv protoc-gen-go

$(PROTOC_GEN_GO_GRPC):
	curl -sSL https://github.com/grpc/grpc-go/releases/download/cmd%2Fprotoc-gen-go-grpc%2Fv$(PROTOC_GEN_GO_GRPC_VERSION)/protoc-gen-go-grpc.v$(PROTOC_GEN_GO_GRPC_VERSION).$(OS).$(ARCH).tar.gz | tar -C $(BIN_DIR) -xzv ./protoc-gen-go-grpc

compile-proto: $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)
	buf generate api

lint-proto:
	buf lint

build-backend-client:
	go build -o client github.com/dtan4/grpc-chat/backend/cmd/client
