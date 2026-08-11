package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/prjvvl/exe/internal/builtin"
	"github.com/prjvvl/exe/internal/config"
	"github.com/prjvvl/exe/internal/dispatch"
	"github.com/prjvvl/exe/internal/registry"
)

type entry struct {
	name        string
	description string
	builtin     *builtin.Builtin
	app         *config.App
}

func Run(reg registry.Registry) error {
	reader := bufio.NewReader(os.Stdin)

	var entries []entry
	for i := range reg.Builtins {
		b := reg.Builtins[i]
		entries = append(entries, entry{name: b.Name, description: b.Description, builtin: &b})
	}
	for i := range reg.Config.Apps {
		a := reg.Config.Apps[i]
		entries = append(entries, entry{name: a.Name, description: a.Description, app: &a})
	}

	for {
		fmt.Println("\nexe: select an application:")
		for i, e := range entries {
			fmt.Printf("  [%d] %-10s %s\n", i+1, e.name, e.description)
		}
		fmt.Println("  [0] quit")
		fmt.Print("> ")

		choice, err := readChoice(reader)
		if err != nil {
			return err
		}
		if choice == "" || choice == "0" {
			return nil
		}

		idx, err := strconv.Atoi(choice)
		if err != nil || idx < 1 || idx > len(entries) {
			fmt.Println("invalid choice")
			continue
		}

		e := entries[idx-1]

		if e.builtin != nil {
			r, err := e.builtin.Handler(reg.Config, nil)
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			return dispatch.Emit(r)
		}

		executed, err := runCommandMenu(reader, *e.app)
		if err != nil {
			return err
		}
		if executed {
			return nil
		}
	}
}

func RunApp(app config.App) error {
	reader := bufio.NewReader(os.Stdin)
	_, err := runCommandMenu(reader, app)
	return err
}

func runCommandMenu(reader *bufio.Reader, app config.App) (bool, error) {
	for {
		fmt.Printf("\nexe %s: select a command:\n", app.Name)
		for i, c := range app.Commands {
			fmt.Printf("  [%d] %-10s %s\n", i+1, c.Name, c.Description)
		}
		fmt.Println("  [0] back")
		fmt.Print("> ")

		choice, err := readChoice(reader)
		if err != nil {
			return false, err
		}
		if choice == "" || choice == "0" {
			return false, nil
		}

		idx, err := strconv.Atoi(choice)
		if err != nil || idx < 1 || idx > len(app.Commands) {
			fmt.Println("invalid choice")
			continue
		}

		cmd := app.Commands[idx-1]
		if err := dispatch.Emit(dispatch.Resolved{Kind: dispatch.KindEval, Text: cmd.Run}); err != nil {
			return false, err
		}
		return true, nil
	}
}

func readChoice(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
