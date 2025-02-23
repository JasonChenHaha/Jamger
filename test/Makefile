SERVER = $(notdir $(CURDIR))

run:
	@echo test...
	@go run ./ ./config.yml

build:
	@echo build...
	go build -o ./test

start:
	@echo start...
	./test ./config.yml

.PHONY: run build start