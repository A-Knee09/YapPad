# Note Maker (Bubble Tea)

A simple terminal-based note-taking app built with Go and Bubble Tea.

## Requirements

- Go 1.21+ (recommended)

## Run with Go

Clone the repository
```bash
https://github.com/A-Knee09/Note-maker-bubbletea.git
cd Note-maker-bubbletea
```
and run:

```bash
go run main.go
```

## Run with Makefile
```bash
make run
```

## Build Binary
```bash 
make build
```

The compiled binary will be created in the project directory.

## Notes Storage

All notes are stored locally in:

$HOME/.notemaker

Each note is saved as a file inside that directory.
