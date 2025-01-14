SERVER = $(notdir $(CURDIR))

build:
	@echo build $(SERVER)...
	@mkdir -p ../../out/$(SERVER)/log
	go build -o ../../out/$(SERVER)/$(SERVER)
	@cp ../../template/config.yml ../../out/$(SERVER)/config.yml

buildraw:
	@echo buildraw...
	@echo $(SERVER)
	@mkdir -p ../../out/$(SERVER)/log
	go build -o ../../out/$(SERVER)/$(SERVER) -gcflags="all=-N -l"

clean:
	@echo clean...
	@rm -rf ../../out/$(SERVER)

.PHONY: build clean