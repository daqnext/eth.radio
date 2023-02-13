

build:
	go build -o ens-server

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ens-server-amd64

PHONY += build

.PHONY: $(PHONY)