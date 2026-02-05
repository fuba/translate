PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
XDG_CONFIG_HOME ?= $(HOME)/.config
CONFIG_DIR ?= $(XDG_CONFIG_HOME)/translate

.PHONY: install
install:
	@mkdir -p "$(BINDIR)"
	@echo "Installing fonts..."
	@./scripts/install-fonts.sh
	@echo "Building translate..."
	@go build -o "$(BINDIR)/translate" ./cmd/translate
	@echo "Installed to $(BINDIR)/translate"
	@if ! echo ":$(PATH):" | grep -q ":$(BINDIR):"; then \
		echo "warning: $(BINDIR) is not in PATH"; \
		echo "Add this to your shell profile:"; \
		echo "  export PATH=\"$(BINDIR):$$PATH\""; \
	fi
