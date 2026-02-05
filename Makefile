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
	@config_dir="$(XDG_CONFIG_HOME)/translate"; \
	config_path="$$config_dir/config.json"; \
	if [ ! -f "$$config_path" ]; then \
		mkdir -p "$$config_dir"; \
		cat > "$$config_path" <<'JSON'; \
{ \
  "base_url": "", \
  "model": "gpt-oss-20b", \
  "from": "", \
  "to": "", \
  "format": "", \
  "timeout_seconds": 120, \
  "max_chars": 2000, \
  "endpoint": "completion", \
  "passphrase_ttl_seconds": 600, \
  "pdf_font": "" \
} \
JSON \
		echo "Wrote config template to $$config_path"; \
		echo "Set base_url in the config to your API endpoint."; \
	else \
		echo "Config exists at $$config_path"; \
	fi
	@if ! echo ":$(PATH):" | grep -q ":$(BINDIR):"; then \
		echo "warning: $(BINDIR) is not in PATH"; \
		echo "Add this to your shell profile:"; \
		echo "  export PATH=\"$(BINDIR):$$PATH\""; \
	fi
