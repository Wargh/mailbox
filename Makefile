.PHONY: build
build:
	./script/build.sh

.PHONY: build-lambda
build-lambda:
	./script/build.sh --zip-only

.PHONY: clean
clean:
	rm -rf ./bin

.PHONY: deploy
deploy: clean build
	sls deploy --verbose

.PHONY: remove
remove: clean
	sls remove --verbose

.PHONY: test
test:
	go test -race -covermode=atomic ./...
