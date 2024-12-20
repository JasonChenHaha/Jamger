ROOT=$(shell pwd)
EXCLUDE=out test project

install:
	@echo install...
	@./script/init_go.sh
	@./script/init_global.sh

clean:
	@echo clean...
	@find ./ \( -name "go.mod" -o -name "go.sum" -o -name "go.work" -o -name "go.work.sum" \) -delete
	@rm -rf out
	@cd ./pb && make -s clean;

build:
	@find ./group -type f -name 'Makefile' -execdir $(MAKE) -s build \;
	@cp ./template/config.yml ./out/config.yml
	@cp ./template/serverList ./out/serverList
	@cp ./template/ctrl.sh ./out/ctrl.sh

run: build
	@./out/ctrl.sh start

start:
	@./out/ctrl.sh start

info:
	@./out/ctrl.sh info

stop:
	@./out/ctrl.sh stop

pb:
	@cd ./pb && make -s;

test:
	@cd ./test && make -s;

.PHONY: install clean build run start stop pb test
