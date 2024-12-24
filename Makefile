ROOT=$(shell pwd)
EXCLUDE=out test project

all:
	@echo Welcome to Jamger world, may your survival be lone, may your death be swift.

install:
	@echo install...
	@./script/init_go.sh
	@./script/init_global.sh

clean:
	@echo clean...
	@find ./ \( -name "go.mod" -o -name "go.sum" -o -name "go.work" -o -name "go.work.sum" \) -delete
	@rm -rf out
	@cd ./pb && make -s clean;

create: 
	@echo create...
	@mkdir ./group/$(p)
	@cp ./template/main.go ./group/$(p)
	@cp ./template/Makefile ./group/$(p)

build:
	@find ./group -type f -name 'Makefile' -execdir $(MAKE) -s build \;
	@cp ./template/config.yml ./out/config.yml
	@cp ./template/serverList ./out/serverList
	@cp ./template/ctrl.sh ./out/ctrl.sh

run: build
	@echo run...
	@./out/ctrl.sh start

start:
	@echo start...
	@./out/ctrl.sh start

info:
	@./out/ctrl.sh info

stop:
	@echo stop...
	@./out/ctrl.sh stop

pb:
	@echo pb...
	@cd ./pb && make -s;

test:
	@echo test...
	@cd ./test && make -s;

.PHONY: install clean build run start stop pb test
