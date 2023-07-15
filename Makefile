VERSION?=0.1.0
COMMIT=$(if $(shell git rev-parse HEAD),$(shell git rev-parse HEAD),"N/A")
DATE=$(shell date "+%Y/%m/%d %H:%M:%S")
LDFLAGS=-ldflags "-s -w -X 'main.buildVersion=$(VERSION)' -X 'main.buildDate=$(DATE)' -X 'main.buildCommit=$(COMMIT)'"

clean:
	@rm -rf ./bin

build:
	go build $(LDFLAGS) -o bin/agent cmd/agent/main.go
	go build $(LDFLAGS) -o bin/server cmd/server/main.go
	go build -o bin/staticlint cmd/staticlint/main.go

test:
	go test -v gometric/internal/...
	
statictest:
	go vet -vettool=$(shell which statictest) ./...

staticlint:
	./bin/staticlint -test=false -G107=false ./cmd/agent/...
	./bin/staticlint -test=false -G107=false ./cmd/server/...
	./bin/staticlint -test=false -G107=false ./internal/...

lint: statictest staticlint