ROOT=$(shell pwd)
EXCLUDE=out test project

all:
	@echo Welcome to Jamger world, may your survival be lone, may your death be swift.

install:
	@echo install...
	@./script/init_go.sh

clean:
	@echo clean...
	@find ./ \( -name "go.mod" -o -name "go.sum" -o -name "go.work" -o -name "go.work.sum" \) -delete
	@rm -rf out
	@cd ./pb && make -s clean;

create: 
	@echo create...
	@mkdir ./server/$(p)
	@cp ./template/main.go ./server/$(p)
	@cp ./template/Makefile ./server/$(p)
	@./script/init_go.sh

build:
	@find ./server -type f -name 'Makefile' -execdir $(MAKE) -s build \;

buildraw:
	@find ./server -type f -name 'Makefile' -execdir $(MAKE) -s buildraw \;

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
	@cd ./test && make -s;

.PHONY: all install clean build run start stop pb test
