# Configuration
BINARY_NAME=dist/cin2ics
MAIN_PACKAGE=main.go

all:
	make test
	make build

build:
	go build -o $(BINARY_NAME)

build-lambda:
	make clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda -o $(BINARY_NAME)_lambda
	zip ${BINARY_NAME}_lambda.zip ${BINARY_NAME}_lambda

test:
	go test -v ./...

clean:
	go clean -v ${MAIN_PACKAGE}
	rm -rf $(BINARY_NAME)
	rm -rf $(BINARY_NAME)_lambda
	rm -rf $(BINARY_NAME)_lambda.zip

run:
	go run ${MAIN_PACKAGE}
