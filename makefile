.PHONY: build run install uninstall clean

build:
	go build -o yap main.go

run:
	go run main.go

install: build
	mkdir -p ~/.local/bin
	cp yap ~/.local/bin/
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
	rm -f yap
	@echo "✓ Build artifacts cleaned."
