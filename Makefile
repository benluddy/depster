CGO_ENABLED := 0
GOFLAGS := -trimpath
GO := go
MODULE := $(shell $(GO) list -m)
VERSION := $(shell git describe --always --tags HEAD)$(and $(shell git status --porcelain),+$(shell scripts/worktree-hash.sh))

.PHONY: all
all: depster

.PHONY: depster
depster:
	$(GO) build -o $@ -ldflags '-X $(MODULE)/internal/version.Version=$(VERSION)' .

.PHONY: clean
clean:
	rm -f depster
