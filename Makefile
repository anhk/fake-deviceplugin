export GOPROXY=https://goproxy.cn,direct
export GO111MODULE=on


.PHONY: all

OBJ := fake-deviceplugin

all: $(OBJ)

$(OBJ):
	go build -mod=vendor -gcflags "-N -l" -o $@ ./

.PHONY: clean
clean:
	rm -fr $(OBJ)

-include .deps

.PHONY: dep
dep:
	echo "$(OBJ): \\" > .deps
	find ./ -path ./vendor -prune -o -name '*.go' -print | awk '{print $$0 " \\"}' >> .deps
	echo "" >> .deps