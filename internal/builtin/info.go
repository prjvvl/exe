package builtin

import (
	"fmt"
	"strings"

	"github.com/prjvvl/exe/internal/config"
	"github.com/prjvvl/exe/internal/dispatch"
	"github.com/prjvvl/exe/internal/version"
)

func InfoBuiltin() Builtin {
	return Builtin{
		Name:        "info",
		Description: "Explain exe (for humans or AI agents) and show the current config",
		Handler: func(cfg *config.Config, args []string) (dispatch.Resolved, error) {
			path, err := config.Path()
			if err != nil {
				return dispatch.Resolved{}, err
			}

			var b strings.Builder

			fmt.Fprintf(&b, "exe %s: a personal command launcher.\n\n", version.Version)
			b.WriteString("How it works: `exe <app> <command>` resolves to a literal shell-command string.\n")
			b.WriteString("exe itself never executes that command. It hands it off to a shell-hook function\n")
			b.WriteString("(loaded in the user's shell profile) which evals it in their actual terminal. This\n")
			b.WriteString("is why `nav`-style commands (e.g. `cd ...`) work: exe is a subprocess and cannot\n")
			b.WriteString("change its parent shell's state on its own.\n\n")
			b.WriteString("Running `exe` with no arguments opens an interactive menu (apps, then commands).\n")
			b.WriteString("Running `exe <app>` jumps straight to that app's command list.\n\n")

			fmt.Fprintf(&b, "Config file: %s\n\n", path)

			b.WriteString("Config format (TOML): a list of apps, each with a list of commands:\n\n")
			b.WriteString("  [[apps]]\n")
			b.WriteString("  name = \"<app name>\"\n")
			b.WriteString("  description = \"<what this app is for>\"\n\n")
			b.WriteString("    [[apps.commands]]\n")
			b.WriteString("    name = \"<command name>\"\n")
			b.WriteString("    description = \"<what this command does>\"\n")
			b.WriteString("    run = \"<literal shell command to eval>\"\n\n")

			fmt.Fprintf(&b, "Reserved names (cannot be used as an app name): %s\n\n", strings.Join(config.ReservedNames, ", "))

			b.WriteString("To add, remove, or change a command: edit the config file directly (it's plain\n")
			b.WriteString("text) following the format above, then save. exe validates the whole file on every\n")
			b.WriteString("load and will fail with a specific error message if something is malformed. Treat\n")
			b.WriteString("that as feedback and fix the file rather than guessing.\n\n")

			b.WriteString("Currently configured:\n")
			if len(cfg.Apps) == 0 {
				b.WriteString("  (no apps configured yet)\n")
			}
			for _, app := range cfg.Apps {
				fmt.Fprintf(&b, "  %s: %s\n", app.Name, app.Description)
				for _, c := range app.Commands {
					fmt.Fprintf(&b, "    exe %s %s: %s (runs: %s)\n", app.Name, c.Name, c.Description, c.Run)
				}
			}

			return dispatch.Resolved{Kind: dispatch.KindPrint, Text: b.String()}, nil
		},
	}
}
