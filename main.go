package main

import (
	"fmt"
	"os"

	"github.com/prjvvl/exe/internal/config"
	"github.com/prjvvl/exe/internal/dispatch"
	"github.com/prjvvl/exe/internal/registry"
	"github.com/prjvvl/exe/internal/shellhook"
	"github.com/prjvvl/exe/internal/tui"
)

func main() {
	args := os.Args[1:]

	// handled before the hook exists, since it's what creates the hook
	if len(args) >= 1 && args[0] == "init" {
		shell := "powershell"
		if len(args) > 1 {
			shell = args[1]
		}
		snippet, err := shellhook.Generate(shell)
		if err != nil {
			fail(err)
		}
		fmt.Println(snippet)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	reg := registry.Build(cfg)

	if len(args) == 0 {
		if err := tui.Run(reg); err != nil {
			fail(err)
		}
		return
	}

	name := args[0]
	rest := args[1:]

	if b, ok := reg.FindBuiltin(name); ok {
		r, err := b.Handler(cfg, rest)
		if err != nil {
			fail(err)
		}
		if err := dispatch.Emit(r); err != nil {
			fail(err)
		}
		return
	}

	app, ok := cfg.FindApp(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "exe: unknown app %q (try `exe info`)\n", name)
		os.Exit(1)
	}

	if len(rest) == 0 {
		if err := tui.RunApp(app); err != nil {
			fail(err)
		}
		return
	}

	cmd, ok := app.FindCommand(rest[0])
	if !ok {
		fmt.Fprintf(os.Stderr, "exe %s: unknown command %q\n", name, rest[0])
		os.Exit(1)
	}

	if err := dispatch.Emit(dispatch.Resolved{Kind: dispatch.KindEval, Text: cmd.Run}); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "exe:", err)
	os.Exit(1)
}
