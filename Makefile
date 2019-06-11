.PHONY: deps vet format test publish clean

SHELL   = bash
GOPATH  ?= $(HOME)/.go
GOBIN   := $(GOPATH)/bin
PATH    := $(GOPATH)/bin:$(PATH)
REPO    := github.com/rabbitt/portunus
PROJECT := portunus
DOCKER_USERNAME ?= user
DOCKER_PASSWORD ?= password

LDFLAGS := -ldflags "-X $(REPO)/$(PROJECT).Revision=$$(git rev-parse HEAD)"

.ONESHELL:

all: vet format test $(PROJECT) publish

all-archs: linux windows darwin

deps:
	@echo "--- collecting ingredients :bento:"
#	go mod download

vet: deps
	@export GOPATH=$(GOPATH)
	go list -f '{{.Dir}}' ./... | grep -vP '(/vendor/|portunus$$)' | xargs go vet -all

format:
	@echo "--- checking for dirty ingredients :mag_right:"
	export GOPATH=$(GOPATH)
	declare -a files
	files=( $$(gofmt -l $$(find . -name '*.go' -a -not -regex '.+/vendor/.+' | xargs)) )
	if [[ $${#files[@]} -ge 1 ]]; then
		[[ $${#files[@]} -eq 1 ]] && s= || s=s
		echo "Found $${#files[@]} dirty ingredient$${s} :face_nose: - cleaning..."
		for ((i = 0; i < $${#files[@]}; i++)); do
			echo -en "\t$${files[$$i]} -> "
			if gofmt -w "$${files[$$i]}" >&-; then
				echo "cleaned :sparkles:"
			else
				echo "still dirty :slightly_frowning_face: - manual intervention required."
			fi
		done
	fi

test: format vet deps
	@echo "+++ Is this thing working? :hammer_and_wrench:"
	GOPATH=$(GOPATH) go test -cover -v

$(PROJECT): deps
	CGO_ENABLED=0 GOPATH=$(GOPATH) go build $(LDFLAGS) -o bin/$@ -v
	touch $@ && chmod 755 $@

linux: deps
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPATH=$(GOPATH) go build $(LDFLAGS) -o bin/$(PROJECT)-linux-amd64 -v
	touch $(PROJECT)-linux-amd64 && chmod 755 $(PROJECT)-linux-amd64

windows: deps
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GOPATH=$(GOPATH) go build $(LDFLAGS) -o bin/$(PROJECT)-windows-amd64.exe -v
	touch $(PROJECT)-windows-amd64.exe

darwin: deps
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 GOPATH=$(GOPATH) go build -o bin/$(PROJECT)-darwin-amd64 -v
	touch $(PROJECT)-darwin-amd64 && chmod 755 $(PROJECT)-darwin-amd64

ifdef TRAVIS_TAG
publish: deps
	@echo "+++ release :octocat:"
	docker login -u "$(DOCKER_USERNAME)" -p "$(DOCKER_PASSWORD)"
	goreleaser --skip-validate --rm-dist
endif

clean:
	rm -rf bin/$(PROJECT) bin/$(PROJECT)-windows-amd64.exe bin/$(PROJECT)-linux-amd64 bin/$(PROJECT)-darwin-amd64 dist
