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
	@cp ./template/config.yml ./server/$(p)
	@./script/init_go.sh

build:
	@find ./server -type f -name 'Makefile' -execdir $(MAKE) -s build \;

buildraw:
	@find ./server -type f -name 'Makefile' -execdir $(MAKE) -s buildraw \;

run: build
	@echo run...
	@find ./out -type f -name 'ctrl.sh' -execdir sh {} start \;

start:
	@echo start...
	@find ./out -type f -name 'ctrl.sh' -execdir sh {} start \;

info:
	@find ./out -type f -name 'ctrl.sh' -execdir sh {} info \;

stop:
	@echo stop...
	@find ./out -type f -name 'ctrl.sh' -execdir sh {} stop \;

pb:
	@echo pb...
	@cd ./pb && make -s;

test:
	@cd ./test && make -s;

.PHONY: all install clean build run start stop pb test
