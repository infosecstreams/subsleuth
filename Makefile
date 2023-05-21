BINARY_NAME=subsleuth
VERSION=0.1.0

all: clean build

clean:
	rm -rf bin/

build:
	go build -a -v -o bin/$(BINARY_NAME)

release:
	go build -a -v -ldflags="-extldflags '-static' -s -w" -tags 'osusergo,netgo,static' -asmflags 'all=-trimpath={{.Env.GOPATH}}' -o bin/$(BINARY_NAME)-$(VERSION)-linux-amd64
