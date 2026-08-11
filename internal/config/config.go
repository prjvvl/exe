package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Command struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Run         string `toml:"run"`
}

type App struct {
	Name        string    `toml:"name"`
	Description string    `toml:"description"`
	Commands    []Command `toml:"commands"`
}

type Config struct {
	Apps []App `toml:"apps"`
}

var ReservedNames = []string{"info", "view", "edit", "init", "update"}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "exe", "config.toml"), nil
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := &Config{Apps: []App{}}
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("creating default config at %s: %w", path, err)
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config at %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config at %s: %w", path, err)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (c *Config) Validate() error {
	seenApps := map[string]bool{}
	for _, a := range c.Apps {
		if a.Name == "" {
			return fmt.Errorf("an app is missing a name")
		}
		for _, r := range ReservedNames {
			if a.Name == r {
				return fmt.Errorf("app name %q is reserved for a built-in command", a.Name)
			}
		}
		if seenApps[a.Name] {
			return fmt.Errorf("duplicate app name %q", a.Name)
		}
		seenApps[a.Name] = true

		seenCmds := map[string]bool{}
		for _, cmd := range a.Commands {
			if cmd.Name == "" {
				return fmt.Errorf("app %q has a command missing a name", a.Name)
			}
			if seenCmds[cmd.Name] {
				return fmt.Errorf("app %q has duplicate command name %q", a.Name, cmd.Name)
			}
			seenCmds[cmd.Name] = true
			if cmd.Run == "" {
				return fmt.Errorf("app %q command %q has no run instruction", a.Name, cmd.Name)
			}
		}
	}
	return nil
}

func (a App) FindCommand(name string) (Command, bool) {
	for _, c := range a.Commands {
		if c.Name == name {
			return c, true
		}
	}
	return Command{}, false
}

func (c *Config) FindApp(name string) (App, bool) {
	for _, a := range c.Apps {
		if a.Name == name {
			return a, true
		}
	}
	return App{}, false
}
