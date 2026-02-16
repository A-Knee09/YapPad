# YapPad
A simple terminal-based note-taking app built with Go and Bubble Tea.

> Built while learning from this YouTube [tutorial](https://youtu.be/pySTZsbaJ0Q?si=5NaxazX5_7UUf19h) to explore TUI development with Bubble Tea.


https://github.com/user-attachments/assets/89aab783-6baf-4028-a88a-b437518bd481


## Requirements
- Go 1.21+ (recommended)

## Installation

Clone the repository:
```bash
git clone https://github.com/A-Knee09/YapPad.git
cd YapPad
```

### Install as a command-line tool
To install YapPad and make it accessible from anywhere with the `yap` command:

```bash
make install
```

This will build the binary and install it to `~/.local/bin/yap`. Make sure `~/.local/bin` is in your PATH.

**For most shells (bash, zsh, fish)**, this directory is already included. If not, add it to your shell config:
- **Bash/Zsh**: Add `export PATH="$HOME/.local/bin:$PATH"` to `~/.bashrc` or `~/.zshrc`
- **Fish**: Run `fish_add_path ~/.local/bin`

After installation, simply run:
```bash
yap
```

### Uninstall
```bash
make uninstall
```

## Development

### Run with Go
```bash
go run main.go
```

### Run with Makefile
```bash
make run
```

### Build Binary
```bash 
make build
```
The compiled binary will be created in the project directory.

## Notes Storage
All notes are stored locally in:
```
$HOME/.YapPad
```
Each note is saved as a Markdown file inside that directory.

## Keyboard Shortcuts
- `ctrl+n` - Create new note
- `ctrl+l` - List all notes
- `ctrl+s` - Save current note
- `ctrl+d` - Delete selected note
- `ctrl+r` - Rename selected note
- `esc` - Go back/Close current view
- `ctrl+c` - Quit application
- `/` or start typing - Filter notes in list view
