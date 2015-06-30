PREFIX = $(DESTDIR)/srv/erl_port
BINDIR = $(PREFIX)/bin
NAME = erl_port
export GOPATH := $(CURDIR):$(CURDIR)/_vendor

all: build

start: build
	./bin/$(NAME)

build:
	go install $(NAME)/...

install: build
	install -D bin/$(NAME) $(BINDIR)/$(NAME)

clean:
	rm -rf pkg bin/$(NAME) _vendor/pkg

test:
	go test ./...

racetest:
	go test -race ./...

vet:
	go vet ./...

doc:
	godoc -http=:6060

.PHONY: build clean test racetest vet doc
