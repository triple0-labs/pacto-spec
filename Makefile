.SHELLFLAGS := -eu -o pipefail -c
SHELL := /bin/bash

.PHONY: test test-go feature-matrix tiny-smoke

test: test-go

test-go:
	@go test ./...

feature-matrix:
	@./bin/pacto-feature-matrix.sh

tiny-smoke:
	@\
	command -v gh >/dev/null || { echo "gh CLI is required"; exit 1; }; \
	command -v curl >/dev/null || { echo "curl is required"; exit 1; }; \
	command -v sha256sum >/dev/null || { echo "sha256sum is required"; exit 1; }; \
	REPO="triple0-labs/pacto-spec"; \
	TAG="$$(gh api "repos/$$REPO/releases/latest" --jq '.tag_name')"; \
	VERSION="$${TAG#v}"; \
	ARTIFACT="pacto_$${VERSION}_linux_amd64.tar.gz"; \
	RUN_DIR="/tmp/pacto-tiny-smoke"; \
	MOCK_DIR="$$RUN_DIR/mock"; \
	BIN_DIR="$$RUN_DIR/bin"; \
	echo "Latest release: $$TAG"; \
	rm -rf "$$RUN_DIR"; \
	mkdir -p "$$RUN_DIR" "$$BIN_DIR" "$$MOCK_DIR"; \
	curl -sS -L -o "$$RUN_DIR/checksums.txt" "https://github.com/$$REPO/releases/download/$$TAG/checksums.txt"; \
	curl -sS -L -o "$$RUN_DIR/$$ARTIFACT" "https://github.com/$$REPO/releases/download/$$TAG/$$ARTIFACT"; \
	( cd "$$RUN_DIR" && sha256sum -c checksums.txt --ignore-missing | grep "$$ARTIFACT: OK" ); \
	tar -xzf "$$RUN_DIR/$$ARTIFACT" -C "$$BIN_DIR"; \
	PCT="$$BIN_DIR/pacto"; \
	"$$PCT" version; \
	printf '# Tiny Mock Project\n' > "$$MOCK_DIR/README.md"; \
	printf '# PACTO\n' > "$$MOCK_DIR/PACTO.md"; \
	printf '# PLAN {{PLAN_NAME}}\n\n## Steps\n- [ ] First step.\n' > "$$MOCK_DIR/PLANTILLA_PACTO_PLAN.md"; \
	mkdir -p "$$MOCK_DIR/current" "$$MOCK_DIR/done" "$$MOCK_DIR/to-implement" "$$MOCK_DIR/outdated"; \
	"$$PCT" status --root "$$MOCK_DIR"; \
	"$$PCT" new to-implement tiny-plan --root "$$MOCK_DIR"; \
	"$$PCT" status --root "$$MOCK_DIR"; \
	echo "tiny-smoke ok: $$MOCK_DIR";
