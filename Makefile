MODULES := . workflow

.PHONY: test-unit test-race test-published vet bench lint

test-unit:
	@set -e; for module in $(MODULES); do \
		echo "==> $$module"; (cd "$$module" && go test ./...); \
	done

test-race:
	@set -e; for module in $(MODULES); do \
		echo "==> $$module"; (cd "$$module" && go test -race ./...); \
	done

test-published:
	@set -e; for module in $(MODULES); do \
		echo "==> $$module"; (cd "$$module" && GOWORK=off go test ./...); \
	done

vet:
	@set -e; for module in $(MODULES); do \
		echo "==> $$module"; (cd "$$module" && go vet ./...); \
	done

bench:
	@set -e; for module in $(MODULES); do \
		echo "==> $$module"; (cd "$$module" && go test -run '^$$' -bench . ./...); \
	done

lint:
	@set -e; for module in $(MODULES); do \
		echo "==> $$module"; (cd "$$module" && golangci-lint run --max-same-issues=0 --max-issues-per-linter=0 ./...); \
	done
