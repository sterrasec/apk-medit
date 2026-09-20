GOCMD=go
GOTEST=$(GOCMD) test -v
GOBUILD=$(GOCMD) build
BINARY_NAME=medit

all: build deploy

# ./cmd only compiles on Linux (ptrace), so keep the default target usable on macOS.
test:
	$(GOTEST) ./pkg/...

test-all:
	$(GOTEST) ./...

build:
	GOOS=linux GOARCH=arm64 GOARM=7 $(GOBUILD) -o $(BINARY_NAME)

build-x86_64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)

build-linux: build-x86_64

clean:
	rm $(BINARY_NAME)

deploy:
ifeq ($(shell adb devices | grep -c 'device$$'), 1)
	$(SHELL) -c "adb push $(BINARY_NAME) /data/local/tmp/$(BINARY_NAME)"
else
	@echo 'Android device is not connected....'
endif
