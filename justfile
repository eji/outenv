default: build

build:
    go build -o outenv .

vet:
    go vet ./...

test:
    go test ./...

install:
    go install .

check: vet test

all: check install
