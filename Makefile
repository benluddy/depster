CGO_ENABLED := 0
GOFLAGS := -trimpath
GO := go
MODULE := $(shell $(GO) list -m)
VERSION := $(shell git describe --always --tags HEAD)$(and $(shell git status --porcelain),+$(shell scripts/worktree-hash.sh))
BINARY := depster$(shell $(GO) env GOEXE)

.PHONY: all
all: $(BINARY)

.PHONY: $(BINARY)
$(BINARY):
	$(GO) build -o $@ -ldflags '-X $(MODULE)/internal/version.Version=$(VERSION)' .

.PHONY: test
unit:
	$(GO) test -coverprofile=coverage.out ./...

.PHONY: clean
clean:
	rm -f depster
