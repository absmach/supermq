BUILD_DIR=build

all: manager http writer coap
.PHONY: all manager http writer coap

define compile_service
	GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -ldflags "-s -w" -o ${BUILD_DIR}/mainflux-$(1) cmd/$(1)/main.go
endef

manager:
	$(call compile_service,$(@))

http:
	$(call compile_service,$(@))

writer:
	$(call compile_service,$(@))


coap:
	$(call compile_service,$(@))

clean:
	rm -rf ${BUILD_DIR}

install:
	cp ${BUILD_DIR}/* $(GOBIN)
