PROJ_NAME = $(notdir $(CURDIR))

all: build

run: build install start

build:
	@echo build...
	@go build
	@mkdir -p ../out/$(PROJ_NAME)
	@cp ../template/* ../out/$(PROJ_NAME)
	@cp config.yml ../out/$(PROJ_NAME)
pb:
	@echo pb...
	@cd ./pb && protoc --go_out=./ *.proto

clean:
	@echo clean...
	@rm -rf ../out/$(PROJ_NAME)

start:
	@echo start...
	@$(OUT_PATH)/start_svr.sh

stop:
	@echo stop...
	@$(OUT_PATH)/stop_svr.sh

test:
	@echo test...
	@cp ./template/config.yml $(TEST_PATH)
	@cd $(TEST_PATH) && go run ./

.PHONY: all install build pb clean start stop test