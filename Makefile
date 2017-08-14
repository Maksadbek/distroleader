all:
		$(MAKE) deps
			$(MAKE) install	

deps:
		go get ./...

install:
		go install ./cmd/distroleader

build:
		go build ./cmd/distroleader

lint:
		go vet ./...

test-all:
		go test ./...

test:
		go test -short ./...
