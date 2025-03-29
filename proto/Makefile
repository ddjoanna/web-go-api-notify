#!make

download:
	@go mod download

install: download
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest


PROTO_PATH := ./notify/v1
PROTO_FILES := $(shell find $(PROTO_PATH) -type f -name '*.proto')
DOC_OUT := ./doc
GO_OUT := ./pkg/$(PROTO_PATH)
OPENAPI_OUT := ./openapi

gen-doc:
	@mkdir -p ${DOC_OUT}
	@protoc --proto_path=${PROTO_PATH} \
		--doc_out=${DOC_OUT} --doc_opt=html,index.html ${PROTO_FILES}
	@protoc --proto_path=${PROTO_PATH} \
		--doc_out=${DOC_OUT} --doc_opt=markdown,README.md ${PROTO_FILES}

# gen swagger docs
gen-go:
	@mkdir -p ${GO_OUT} ${OPENAPI_OUT}
	@protoc --proto_path=${PROTO_PATH} \
		--go_out=${GO_OUT} \
		--go_opt=paths=source_relative \
		--go-grpc_out=${GO_OUT} \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=${GO_OUT} \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		--openapiv2_out=${OPENAPI_OUT} \
		--openapiv2_opt=generate_unbound_methods=true \
		${PROTO_FILES}


gen: gen-doc gen-go