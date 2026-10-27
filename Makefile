AUTH_USER    ?=
AUTH_PASS    ?=
FILE         ?= tests/testdata/m600-gps.json
NEXT_VERSION := $(shell svu next --v0 --always)

.PHONY: build run test-certs mock-api release test lint vet clean

build:
	go build -buildmode=pie -o meinberg_ltos_exporter .

run: build
	./meinberg_ltos_exporter --target https://localhost:8080 --web.listen-address localhost:10123 --ignore-ssl-verify $(if $(AUTH_USER),--auth-user $(AUTH_USER)) $(if $(AUTH_PASS),--auth-pass $(AUTH_PASS))

test-certs:
	 openssl req -x509 -newkey rsa:4096 -keyout tests/mock-key.pem -out tests/mock-cert.pem -sha256 -days 5 -nodes -subj "/C=CH/ST=State/L=City/O=Organization/OU=Department/CN=localhost"

mock-api: test-certs
	go run tests/mock-server.go -ssl-cert tests/mock-cert.pem -ssl-key tests/mock-key.pem -file $(FILE) $(if $(AUTH_USER),-user $(AUTH_USER)) $(if $(AUTH_PASS),-pass $(AUTH_PASS))

release:
	@echo "Current version: $(shell svu current)"
	@echo "Next version:    $(NEXT_VERSION)"
	@read -p "Press enter to confirm release or Ctrl+C to cancel"
	git tag -a $(NEXT_VERSION) -m "Release $(NEXT_VERSION)"
	git push origin $(NEXT_VERSION)

test:
	go test -race -cover -v ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

clean:
	@echo "Removing build artifacts, generated files, and test certificates..."
	rm -rfv dist/ meinberg_ltos_exporter tests/*.pem
