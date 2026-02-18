.PHONY: build run install uninstall clean

build:
	mkdir -p build/
	go build -o build/yap .

run:
	go run .

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
