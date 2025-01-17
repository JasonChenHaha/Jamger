PROJ_NAME = $(notdir $(CURDIR))

all:
	@echo pb...
	@gopath=$(shell go env GOPATH); \
	protoc --go_out=. ./*.proto

clean:
	@rm -f *.pb.go

.PHONY: all clean