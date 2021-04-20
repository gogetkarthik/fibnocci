.PHONY: clean
TOOLS := \
	github.com/mitchellh/gox;

GO     := $(shell command -v go 2> /dev/null)
DOCKER := $(shell command -v docker 2> /dev/null)

GETTOOLS  := $(foreach TOOL, ${TOOLS}, go get ${TOOL})

.PHONY: check-go-installed

check-go-installed:
ifndef GO
	$(error "go is not installed, please install before proceeding")
endif
	echo "go installed"

docker-build-local:
ifndef DOCKER
	$(error "docker is not installed please install before proceeding")
endif
	docker build --no-cache -f Dockerfile -t  karthikeyan2418/fibonacci:latest .

tools: check-go-installed
	$(GETTOOLS)

dist: check-go-installed
	mkdir -p dist
	gox -verbose -os="darwin linux" -arch="amd64" -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" ./cmd/fibonacci/

test: check-go-installed
	go test ./...

vet: check-go-installed
	go vet ./...

fmt: check-go-installed
	go fmt ./...

build:
	go build ./...

clean:
	rm -rf dist
#default: tools vet fmt
