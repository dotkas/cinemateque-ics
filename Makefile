# Configuration
BINARY_NAME=.build/cin2ics
MAIN_PACKAGE=main.go

all:
	make test
	make build

build:
	go build -o $(BINARY_NAME)

build-lambda:
	make clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build

test:
	go test -v ./...

clean:
	go clean -v ${MAIN_PACKAGE}
	rm -rf $(BINARY_NAME)

run:
	go run ${MAIN_PACKAGE}
