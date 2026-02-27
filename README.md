# YapPad

A terminal-based note-taking and journaling app built with Go and Bubble Tea.

> Built while learning from this YouTube [tutorial](https://youtu.be/pySTZsbaJ0Q?si=5NaxazX5_7UUf19h) to explore TUI development with [BubbleTea](https://github.com/charmbracelet/bubbletea)

## Showcase


https://github.com/user-attachments/assets/353620d4-dcfe-40d5-ada4-224adad019ad

### Image rendering
<img width="1289" height="817" alt="YapPad image rendering" src="https://github.com/user-attachments/assets/93507e04-aa27-43e4-adde-02299d89090a" />



> [!IMPORTANT]
> - Tested only on Linux as of now
> - Image preview requires a Kitty-compatible terminal (e.g. Kitty, WezTerm) and `chafa` installed. It will not work in standard terminals like GNOME Terminal, Alacritty, or tmux.
> - Still in development, bugs are expected. All the above mentioned will also be fixed 

## Requirements

- Go 1.21+
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

- Bash/Zsh: `export PATH="$HOME/.local/bin:$PATH"`
- Fish: `fish_add_path ~/.local/bin`

### Uninstall

```bash
make uninstall
```

## Usage

```bash
yap                        # Open default vault (~/.YapPad) in "all" mode
yap .                      # Open in current directory (WIP)
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

### Creating Notes

Press `ctrl+n` to enter creation mode. You will be prompted for a filename first, then an optional description. Pressing enter on an empty filename auto-generates a date-stamped file in the current mode's directory (e.g. `daily/2026-02-18.md`). Press `tab` while typing the filename to cycle through journal modes before creating. Pressing enter on an empty description skips it and falls back to showing the modified date.

### Descriptions

Each note can have a custom description that appears in the file list beneath its title. Descriptions are stored in `~/.YapPad/.metadesc/` as hidden sidecar files and do not modify the note content at all. If no description is set, the last modified date is shown instead.

### Renaming Notes

Press `ctrl+r` to rename the selected note. You will be prompted for the new filename and then a new description. If you skip the description step, the existing description is preserved automatically.

### Templates

Place template files in `~/.YapPad/.templates/` named after the mode (`daily.md`, `weekly.md`, etc.). New default journal entries will be pre-filled with the matching template content.

### Preview Pane

Toggle with `ctrl+p`. Displays syntax-highlighted text previews for markdown and code files, and inline image previews for supported image formats. The preview pane auto-hides if the terminal is too narrow (below 80 columns). Image previews require `chafa` and a Kitty-compatible terminal.

### Sorting

Press `ctrl+s` to cycle through sort modes: Modified (newest/oldest), Created (newest/oldest), and Alphabetic (ascending/descending).

### Mouse Support

Scroll the file list and preview pane independently using the mouse wheel.

### Filtering

Press `/` to filter notes by filename within the current view.

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `ctrl+n` | Create new note |
| `ctrl+r` | Rename selected note |
| `ctrl+d` | Delete selected note |
| `ctrl+p` | Toggle preview pane |
| `ctrl+s` | Cycle sort mode |
| `enter` | Open selected note in `$EDITOR` (default: nvim) |
| `0-4` | Switch mode (0=all, 1=daily, 2=weekly, 3=monthly, 4=yearly) |
| `tab` | Cycle journal mode while creating a note |
| `/` | Filter notes by name |
| `?` | Toggle help menu |
| `esc` | Cancel current action |

## Notes Storage

All notes are stored locally in `~/.YapPad/` (or the vault directory you specify). Each note is a plain Markdown file. Descriptions are stored separately in `~/.YapPad/.metadesc/` and do not affect the files themselves.

```
~/.YapPad/
├── daily/
├── weekly/
├── monthly/
├── yearly/
├── .metadesc/
└── .templates/
```

## Development

```bash
make run     # Run with Go
make build   # Build binary to build/yap
make clean   # Remove build artifacts
```
