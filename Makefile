ROOT=$(shell pwd)
EXCLUDE=out test project

install:
	@echo install...
	@./script/init_go.sh
	@./script/init_global.sh

clean:
	@echo clean...
	@find ./ \( -name "go.mod" -o -name "go.sum" -o -name "go.work" -o -name "go.work.sum" \) -delete
	@find ./project -type f -name 'Makefile' -execdir $(MAKE) -s clean \;
	@rm -rf out
	@cd ./pb && make -s clean;

build:
	@find ./project -type f -name 'Makefile' -execdir $(MAKE) -s build \;

run:
	@find ./project -type f -name 'Makefile' -execdir $(MAKE) -s run \;

start:
	@find ./project -type f -name 'Makefile' -execdir $(MAKE) -s start \;

info:
	@find ./project -type f -name 'Makefile' -execdir $(MAKE) -s info \;

stop:
	@find ./project -type f -name 'Makefile' -execdir $(MAKE) -s stop \;

pb:
	@cd ./pb && make -s;

test:
	@cd ./test && make -s;

.PHONY: install clean build run start stop pb test
