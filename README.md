# YapPad

A terminal-based note-taking app built with Go and Bubble Tea.

> Built while learning TUI development with [BubbleTea](https://github.com/charmbracelet/bubbletea)

## Showcase

https://github.com/user-attachments/assets/c601849c-5179-4787-9b64-93ca44c7f397

## Preview

<img width="1588" height="815" alt="MD render" src="https://github.com/user-attachments/assets/22e89bb2-56a3-4057-88ca-f342b9a2c35e" />

<img width="1278" height="748" alt="Preview" src="https://github.com/user-attachments/assets/56d7f0c6-4223-4ad9-a9cc-be301fac1aa3" />

> [!IMPORTANT]
> - Tested only on Linux
> - Image preview requires a Kitty-compatible terminal (Kitty, WezTerm) and `chafa` installed
> - Image preview might take some time to load initially or on terminal resize, but not again

## Requirements

- Go 1.21+
- [chafa](https://hpjansson.org/chafa/) (optional, for image previews)

## Installation

```bash
git clone https://github.com/A-Knee09/YapPad.git
cd YapPad
make install
```

Installs to `~/.local/bin/yap`. Make sure it's in your PATH:

```bash
# Bash/Zsh
export PATH="$HOME/.local/bin:$PATH"

# Fish
fish_add_path ~/.local/bin
```

```bash
make uninstall  # to remove
```

## First Run

On first run, YapPad will ask you to configure:

- Where to store your notes (default: `~/.YapPad`)
- Which editor to use (default: `inbuilt`)
- Which theme to use (default: `default`)

Config is saved to `~/.config/yappad/config.toml` and can be edited directly.

## Usage

```bash
yap                   # open your configured vault
yap .                 # open current directory as vault (this session only)
yap --theme gruvbox   # override theme for this session
yap --editor nvim     # override editor for this session
yap --version         # print version
yap --help            # show help
```

## Features

### Notes & Files

Press `n` to create a new note. You will be prompted for a filename, then an optional description. Pressing enter on an empty filename creates a date-stamped file. Pressing enter on an empty description skips it and falls back to showing the modified date.

Each note can have a custom description shown below its title in the list. Descriptions are stored in `.metadesc/` as sidecar files and never modify note content.

Press `r` to rename a note. You will be prompted for the new name and description. Skipping the description preserves the existing one.

### Preview Pane

Toggle with `ctrl+p`. Shows syntax-highlighted text and markdown previews, and inline image previews for supported formats. Auto-hides if the terminal is too narrow. Image preview requires `chafa` and a Kitty-compatible terminal.

### Themes

16 built-in themes: `default`, `gruvbox`, `nord`, `tokyonight`, `forest`, `solarized`, `dracula`, `dusk`, `tide`, `moss`, `glacier`, `plum`, `algae`, `sunny`, `stone`.

Set in config or override per session with `--theme <name>`.

### Sorting

Press `ctrl+s` to cycle through sort modes: Modified (newest/oldest), Created (newest/oldest), Alphabetic (ascending/descending).

### Editors

Set `editor` in config or pass `--editor` flag. Supports `inbuilt`, `nano`, `nvim`, `vim`, `hx`, or any editor in your `$PATH`. The inbuilt editor supports `ctrl+s` to save and `ctrl+q` to close.

### Mouse Support

Scroll the file list and preview pane independently with the mouse wheel.

## Keybindings

| Key | Action |
|-----|--------|
| `ctrl n` | New note |
| `ctrl r` | Rename note |
| `ctrl d` | Delete note |
| `enter` | Open in editor |
| `ctrl+p` | Toggle preview |
| `ctrl+s` | Cycle sort |
| `/` | Filter notes |
| `?` | Toggle help |
| `esc` | Cancel |
| `q` | Quit |

## Config

`~/.config/yappad/config.toml`:

```toml
theme = "default"
editor = "inbuilt"
vault = "/home/user/.YapPad"
```

## Storage

```
~/.YapPad/
├── .metadesc/     # note descriptions
└── .templates/    # optional note templates
```

Notes are plain files. Any file type is supported — markdown, text, images, code files.

## Development

```bash
make run     # run directly
make build   # build to build/yap
make clean   # remove build artifacts
```
