PROJ_NAME = $(notdir $(CURDIR))

all:
	@echo pb...
	@protoc --go_out=./ *.proto

clean:
	@rm -f *.pb.go

.PHONY: all clean