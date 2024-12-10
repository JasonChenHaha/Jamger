PROJ_NAME = $(notdir $(CURDIR))

all:

run: build start

build:
	@echo build...
	@echo $(PROJ_NAME)
	@mkdir -p ../../out/$(PROJ_NAME)
	@cp ../../template/* ../../out/$(PROJ_NAME)
	@sed -i 's/jamger/$(PROJ_NAME)/g' ../../out/$(PROJ_NAME)/start_svr.sh
	@sed -i 's/jamger/$(PROJ_NAME)/g' ../../out/$(PROJ_NAME)/stop_svr.sh
	@cp config.yml ../../out/$(PROJ_NAME)
	@go build -o ../../out/$(PROJ_NAME)

pb:
	@echo pb...
	@cd ./pb && protoc --go_out=./ *.proto

clean:
	@echo clean...
	@rm -rf ../../out/$(PROJ_NAME)

start:
	@echo start...
	@../../out/$(PROJ_NAME)/start_svr.sh

stop:
	@echo stop...
	@../../out/$(PROJ_NAME)/stop_svr.sh

.PHONY: all build clean start stop