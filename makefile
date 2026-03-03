.PHONY: build run install uninstall clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
LDFLAGS := -ldflags="-X 'main.Version=$(VERSION)'"

build:
	mkdir -p build/
	go build $(LDFLAGS) -o build/yap .

run:
	go run $(LDFLAGS) .

install: build
	mkdir -p ~/.local/bin
	cp build/yap ~/.local/bin/
	chmod +x ~/.local/bin/yap
	@echo "✓ YapPad installed successfully!"
	@echo "Run 'yap' from anywhere to start taking notes."
	@echo ""
	@echo "If 'yap' command is not found, make sure ~/.local/bin is in your PATH:"
	@echo "  Bash/Zsh: export PATH=\"\$$HOME/.local/bin:\$$PATH\""
	@echo "  Fish: fish_add_path ~/.local/bin"

uninstall:
	rm -f ~/.local/bin/yap
	@echo "✓ YapPad uninstalled."

clean:
	rm -rf build/
	@echo "✓ Build artifacts cleaned."
