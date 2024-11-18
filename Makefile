SVR_NAME = jamger
OUT_PATH = ./out
TEST_PATH = ./test

all: install build

run: install build start

install:
ifneq ($(shell test -f go.mod && echo 1 || echo 0), 1)
	go mod init $(SVR_NAME)
endif
	go mod tidy
	mkdir -p $(OUT_PATH)
	cp ./template/start_svr.sh $(OUT_PATH); chmod +x $(OUT_PATH)/start_svr.sh; sed -i 's/svr_name/$(SVR_NAME)/' $(OUT_PATH)/start_svr.sh
	cp ./template/stop_svr.sh $(OUT_PATH); chmod +x $(OUT_PATH)/stop_svr.sh; sed -i 's/svr_name/$(SVR_NAME)/' $(OUT_PATH)/stop_svr.sh
	cp ./template/config.yml $(OUT_PATH)

build: install
	make install
	go build -o $(OUT_PATH)

clean:
	rm -rf $(OUT_PATH)

start:
	$(OUT_PATH)/start_svr.sh

stop:
	$(OUT_PATH)/stop_svr.sh

test:
	cp ./template/config.yml $(TEST_PATH)
	cd $(TEST_PATH) && go run ./

.PHONY: all install build clean start stop test