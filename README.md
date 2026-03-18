# YapPad

A terminal-based note-taking app built with Go and Bubble Tea.

> Built while learning TUI development with [BubbleTea](https://github.com/charmbracelet/bubbletea)

## Showcase

https://github.com/user-attachments/assets/cbe425a2-4a54-4b5e-a331-47501feb353d

> [!IMPORTANT]
> - Tested only on Linux and MacOS
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


### Themes
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/cf25288b-df85-4d16-92d1-8901aacac6d3" alt="Default" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/b32972f7-2489-4ab1-9bf8-f23506ce01b4" alt="Dracula" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Default</td>
    <td align="center">Dracula</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/d1445109-5337-4dbd-949f-6f1f1ca83e2c" alt="Dusk" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/8aa1c819-3deb-4d0c-858b-c75a0f01865c" alt="Forest" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Dusk</td>
    <td align="center">Forest</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/d6fa6c25-0c28-4d04-bcd4-fd466e6d3b6a" alt="Glacier" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/7a1b0049-8f8b-4cea-83e2-3f4a6e66a2ff" alt="Gruvbox" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Glacier</td>
    <td align="center">Gruvbox</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/1e25cae9-64f8-41ba-b765-ea23bafe4c3e" alt="Moss" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/a9893933-8962-4c21-902e-cec55d77eaa1" alt="Nord" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Moss</td>
    <td align="center">Nord</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/c0d0328c-144d-4634-af23-f5a10d882b9d" alt="Plum" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/5cd6360e-f2ff-4aa5-89d5-a0190d9434c2" alt="Solarized" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Plum</td>
    <td align="center">Solarized</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/67e5a2c8-3037-40a7-ab8c-dd98a39d0461" alt="Tide" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/ffec77c5-e23c-4335-8262-085fbad83276" alt="Tokyonight" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Tide</td>
    <td align="center">Tokyonight</td>
  </tr>
</table>


16 built-in themes: `default`, `gruvbox`, `nord`, `tokyonight`, `forest`, `solarized`, `dracula`, `dusk`, `tide`, `moss`, `glacier`, `plum`, `algae`, `sunny`, `stone`.

Set in config or override per session with `--theme <name>`.

### Preview Pane

Toggle with `ctrl+p`. Shows syntax-highlighted text and markdown previews, and inline image previews for supported formats. Auto-hides if the terminal is too narrow. Image preview requires `chafa` and a Kitty-compatible terminal.


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
