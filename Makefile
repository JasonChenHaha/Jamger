SVR_NAME = jamger
OUT_PATH = ./out
TEST_PATH = ./test

all: install build

run: install build start

install:
	@echo install...
ifneq ($(shell test -f go.mod && echo 1 || echo 0), 1)
	@go mod init $(SVR_NAME)
endif
	@go mod tidy
	@mkdir -p $(OUT_PATH)
	@cp ./template/start_svr.sh $(OUT_PATH); chmod +x $(OUT_PATH)/start_svr.sh; sed -i 's/svr_name/$(SVR_NAME)/' $(OUT_PATH)/start_svr.sh
	@cp ./template/stop_svr.sh $(OUT_PATH); chmod +x $(OUT_PATH)/stop_svr.sh; sed -i 's/svr_name/$(SVR_NAME)/' $(OUT_PATH)/stop_svr.sh
	@cp ./template/config.yml $(OUT_PATH)

build: install
	@echo build...
	@make install
	@go build -o $(OUT_PATH)

pb:
	@echo pb...
	@cd ./pb && protoc --go_out=./ *.proto

clean:
	@echo clean...
	@rm -rf ./pb/*.pb.go
	@rm -rf $(OUT_PATH)
	@rm -rf ./test/config.yml

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