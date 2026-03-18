// NOTE: Config file loading and first-run setup

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Theme  string `toml:"theme"`
	Editor string `toml:"editor"`
	Vault  string `toml:"vault"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "yappad", "config.toml")
}

func loadConfig() Config {
	cfg := Config{
		Theme:  "default",
		Editor: "",
		Vault:  filepath.Join(os.Getenv("HOME"), ".YapPad"),
	}

	path := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not parse config: %v\n", err)
	}

	if cfg.Vault != "" {
		cfg.Vault = os.ExpandEnv(cfg.Vault)
		if strings.HasPrefix(cfg.Vault, "~") {
			home, _ := os.UserHomeDir()
			cfg.Vault = filepath.Join(home, cfg.Vault[1:])
		}
	}

	return cfg
}

func writeConfig(cfg Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func runSetup() Config {
	reader := bufio.NewReader(os.Stdin)
	cfg := Config{}

	home, _ := os.UserHomeDir()
	defaultVault := filepath.Join(home, ".YapPad")

	fmt.Println("Welcome to YapPad! Let's set things up.\n")

	// Vault
	fmt.Printf("Where do you want to store your notes? (Default directory [%s]) : ", defaultVault)
	vault, _ := reader.ReadString('\n')
	vault = strings.TrimSpace(vault)
	if vault == "" {
		vault = defaultVault
	} else {
		if strings.HasPrefix(vault, "~") {
			vault = filepath.Join(home, vault[1:])
		}
		vault = os.ExpandEnv(vault)
	}
	cfg.Vault = vault

	// Editor
	validEditors := map[string]bool{
		"inbuilt": true, "nano": true, "nvim": true, "vim": true, "hx": true,
	}
	for {
		fmt.Print("Which editor? (inbuilt, nano, nvim, vim, hx) [inbuilt]: ")
		editor, _ := reader.ReadString('\n')
		editor = strings.TrimSpace(editor)
		if editor == "" {
			cfg.Editor = "inbuilt"
			break
		}
		if validEditors[editor] {
			cfg.Editor = editor
			break
		}
		fmt.Printf("  unknown editor %q, try again\n", editor)
	}

	// Theme
	validThemes := map[string]bool{
		"default": true, "gruvbox": true, "nord": true, "tokyonight": true,
		"forest": true, "solarized": true, "dracula": true,
		"dusk": true, "tide": true, "moss": true, "glacier": true,
		"plum": true, "algae": true, "sunny": true, "stone": true,
	}
	for {
		fmt.Print("Which theme? (default, gruvbox, nord, tokyonight, forest, solarized, catppuccin, dracula, dusk, tide, moss, glacier, plum, algae, sunny, stone) [default]: ")
		theme, _ := reader.ReadString('\n')
		theme = strings.TrimSpace(theme)
		if theme == "" {
			cfg.Theme = "default"
			break
		}
		if validThemes[theme] {
			cfg.Theme = theme
			break
		}
		fmt.Printf("  unknown theme %q, try again\n", theme)
	}

	if err := writeConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
	} else {
		fmt.Printf("\nConfig saved to %s\n\n", configPath())
	}

	return cfg
}
