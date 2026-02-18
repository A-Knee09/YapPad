# YapPad
A simple terminal-based note-taking app built with Go and Bubble Tea.

> Built while learning from this YouTube [tutorial](https://youtu.be/pySTZsbaJ0Q?si=5NaxazX5_7UUf19h) to explore TUI development with Bubble Tea.


https://github.com/user-attachments/assets/89aab783-6baf-4028-a88a-b437518bd481


## Requirements
- Go 1.21+ (recommended)
- [chafa](https://hpjansson.org/chafa/) (optional, for image previews in Kitty-compatible terminals)

## Installation

Clone the repository:
```bash
git clone https://github.com/A-Knee09/YapPad.git
cd YapPad
```

### Install as a command-line tool
```bash
make install
```

This installs to `~/.local/bin/yap`. Make sure it's in your PATH:
- **Bash/Zsh**: `export PATH="$HOME/.local/bin:$PATH"`
- **Fish**: `fish_add_path ~/.local/bin`

### Uninstall
```bash
make uninstall
```

## Usage

```bash
yap                        # Open default vault (~/.YapPad) in "all" mode
yap --mode daily           # Open in daily journal mode
yap --mode weekly ~/notes  # Use ~/notes as vault, weekly mode
```

### CLI Options
| Flag | Description |
|------|-------------|
| `--mode <mode>` | Set default yap mode: `all`, `daily`, `weekly`, `monthly`, `yearly` |
| `[vault-dir]` | Optional path to notes directory (default: `~/.YapPad`) |

## Features

### Journal Modes
Notes are organized into subdirectories by frequency: `daily/`, `weekly/`, `monthly/`, `yearly/`. Press `0-4` to switch between All/Daily/Weekly/Monthly/Yearly views.

Creating a note with `ctrl+n` → `Enter` (no name) auto-generates a date-stamped file in the current mode's directory (e.g. `daily/2026-02-18.md`). Use `tab` while in the input to cycle modes before creating.

### Templates
Place template files in `~/.YapPad/.templates/` named after the mode (`daily.md`, `weekly.md`, etc.). New default journal entries will be pre-filled with the matching template.

### Preview Pane
Toggle with `ctrl+p`. Shows syntax-highlighted text previews and inline image previews (requires `chafa` + a Kitty-compatible terminal).

### Mouse Support
Scroll the file list and preview pane with the mouse wheel.

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `ctrl+n` | Create new note |
| `ctrl+r` | Rename selected note |
| `ctrl+d` | Delete selected note |
| `ctrl+p` | Toggle preview pane |
| `ctrl+s` | Cycle sort mode (modified/created, asc/desc) |
| `ctrl+c` | Quit |
| `0-4` | Switch yap mode (0=all, 1=daily, 2=weekly, 3=monthly, 4=yearly) |
| `tab` | Cycle yap mode (while creating a note) |
| `enter` | Open selected note in `$EDITOR` (default: nvim) |
| `/` | Filter notes |
| `esc` | Cancel current action |

## Notes Storage
All notes are stored locally in `~/.YapPad/` (or the vault directory you specify). Each note is a Markdown file.

## Development

```bash
make run     # Run with Go
make build   # Build binary to build/yap
make clean   # Remove build artifacts
```
