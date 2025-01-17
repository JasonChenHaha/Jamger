SERVER = $(notdir $(CURDIR))

build:
	@echo build $(SERVER)...
	@mkdir -p ../../out/$(SERVER)/log
	go build -o ../../out/$(SERVER)/$(SERVER)
	@cp ./config.yml ../../out/$(SERVER)
	@cp ../../template/ctrl.sh ../../out/$(SERVER)
	@cp ../../template/server.crt ../../out/$(SERVER)
	@cp ../../template/server.key ../../out/$(SERVER)

buildraw:
	@echo buildraw...
	@echo $(SERVER)
	@mkdir -p ../../out/$(SERVER)/log
	go build -o ../../out/$(SERVER)/$(SERVER) -gcflags="all=-N -l"
	@cp ./config.yml ../../out/$(SERVER)
	@cp ../../template/ctrl.sh ../../out/$(SERVER)
	@cp ../../template/server.crt ../../out/$(SERVER)
	@cp ../../template/server.key ../../out/$(SERVER)

clean:
	@echo clean...
	@rm -rf ../../out/$(SERVER)

.PHONY: build clean