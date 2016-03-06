APPNAME = marathon-alerts
VERSION=0.0.1-dev
TESTFLAGS=-v

build:
	go build -tags netgo -ldflags "-w" -o ${APPNAME} .

build-linux:
	GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags "-w -s -X main.APP_VERSION=${VERSION}" -v -o ${APPNAME}-linux-amd64 .

build-mac:
	GOOS=darwin GOARCH=amd64 go build -tags netgo -ldflags "-w -s -X main.APP_VERSION=${VERSION}" -v -o ${APPNAME}-darwin-amd64 .

build-all: build-mac build-linux

all: setup
	build
	install

setup:
	glide install

test-only:
	go test ${TESTFLAGS} github.com/ashwanthkumar/marathon-alerts/${name}

test:
	go test ${TESTFLAGS} github.com/ashwanthkumar/marathon-alerts/
