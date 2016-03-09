APPNAME = marathon-alerts
VERSION=0.0.1-dev
TESTFLAGS=-v -cover -coverprofile cover.out
TEST_COVERAGE_THRESHOLD=48.0

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

test-ci: test
	@go tool cover -html=cover.out -o coverage.html
	@go tool cover -func=cover.out | grep "total:" | awk '{print $$3}' | sed -e 's/%//' > coverage.txt
	@bash -c 'COVERAGE=$$(cat coverage.txt);	\
	echo "Current Coverage % is $$COVERAGE, expected is ${TEST_COVERAGE_THRESHOLD}.";	\
	exit $$(echo $$COVERAGE"<${TEST_COVERAGE_THRESHOLD}" | bc -l)'
