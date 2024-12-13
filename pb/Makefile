PROJ_NAME = $(notdir $(CURDIR))

all:
	@echo pb...
	@gopath=$(shell go env GOPATH); \
	protoc --go_out=. --go-grpc_out=. ./*.proto --plugin=$$gopath/bin/protoc-gen-go-grpc

clean:
	@rm -f *.pb.go

.PHONY: all clean