PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin

.PHONY: build test clean install uninstall

build:
	go build -o roster ./cmd/roster/

test:
	go test -v ./...

install: build
	mkdir -p $(BINDIR)
	cp roster $(BINDIR)/roster

uninstall:
	rm -f $(BINDIR)/roster

clean:
	rm -f roster
	rm -rf test_inv
