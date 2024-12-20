GROUP = $(notdir $(CURDIR))

build:
	@echo build...
	@echo $(GROUP)
	@mkdir -p ../../out/$(GROUP)/log
	go build -o ../../out/$(GROUP)/$(GROUP)

clean:
	@echo clean...
	@rm -rf ../../out/$(GROUP)

.PHONY: build clean