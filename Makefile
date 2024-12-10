ROOT=$(shell pwd)
EXCLUDE=out test project

all:

install:
	@echo install...
	@./script/init_go.sh

other:
	# @dirs=""; \
	# for dir in $(shell find . -type d -not -path '*/.*' -not -path '.'); do \
	# 	exclude=0; \
	# 	for ex in $(EXCLUDE); do \
	# 		if echo "$$dir" | grep -q "$$ex"; then \
	# 			exclude=1; \
	# 			break; \
	# 		fi; \
	# 	done; \
	# 	if [[ $$exclude -eq 0 ]]; then \
	# 		cd $(ROOT)/$(patsubst ./%,%,$$dir); \
	# 		if [[ ! -f ./go.mod ]]; then \
	# 			go mod init "j$$(basename $$dir)"; \
	# 		fi; \
	# 		go mod tidy; \
	# 	fi; \
	# done; \
	# for dir in $(shell find ./project -type d -not -path '*/.*' -not -path '.' -not -path './project'); do \
	# 	cd $(ROOT)/$(patsubst ./%,%,$$dir); \
	# 	if [[ $$(echo $$dir | tr -cd '/' | wc -c) -eq 2 ]]; then \
	# 		project=$$(basename $$dir); \
	# 		go mod init "$$project"; \
	# 	else \
	# 		go mod init "$$project$$(basename $$dir)"; \
	# 	fi; \
	# done;

clean:
	@echo clean...
	@find ./ \( -name "go.mod" -o -name "go.sum" -o -name "go.work" -o -name "go.work.sum" \) -delete
	@rm -rf out

.PHONY: all install clean
