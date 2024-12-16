PROJ_NAME = $(notdir $(CURDIR))

all:

run: build start

build:
	@echo build...
	@echo $(PROJ_NAME)
	@mkdir -p ../../out/$(PROJ_NAME)
	@cp ../../template/start_svr.sh ../../out/$(PROJ_NAME)
	@cp ../../template/stop_svr.sh ../../out/$(PROJ_NAME)
	@sed -i 's/jamger/$(PROJ_NAME)/g' ../../out/$(PROJ_NAME)/start_svr.sh
	@sed -i 's/jamger/$(PROJ_NAME)/g' ../../out/$(PROJ_NAME)/stop_svr.sh
	@cp config.yml ../../out/$(PROJ_NAME)
	go build -o ../../out/$(PROJ_NAME)/$(PROJ_NAME)

clean:
	@echo clean...
	@rm -rf ../../out/$(PROJ_NAME)

start:
	@echo start...
	@../../out/$(PROJ_NAME)/start_svr.sh -s

info:
	@ps ux | grep './$(PROJ_NAME)' | grep -v grep || true

stop:
	@echo stop...
	@../../out/$(PROJ_NAME)/stop_svr.sh -s

.PHONY: all build clean start info stop