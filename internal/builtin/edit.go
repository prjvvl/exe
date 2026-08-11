package builtin

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/prjvvl/exe/internal/config"
	"github.com/prjvvl/exe/internal/dispatch"
)

func EditBuiltin() Builtin {
	return Builtin{
		Name:        "edit",
		Description: "Open the config file in your editor",
		Handler: func(cfg *config.Config, args []string) (dispatch.Resolved, error) {
			path, err := config.Path()
			if err != nil {
				return dispatch.Resolved{}, err
			}
			editor := editorCommand()
			// The quoted path is eval'd as-is by the shell hook; a stray
			// double quote in either half would break out of that quoting.
			if strings.Contains(editor, "\"") || strings.Contains(path, "\"") {
				return dispatch.Resolved{}, fmt.Errorf("editor command or config path contains a double quote, refusing to build an unsafe command")
			}
			cmd := fmt.Sprintf("%s \"%s\"", editor, path)
			return dispatch.Resolved{Kind: dispatch.KindEval, Text: cmd}, nil
		},
	}
}

func editorCommand() string {
	if e := os.Getenv("EXE_EDITOR"); e != "" {
		return e
	}
	if e := os.Getenv("VISUAL"); e != "" {
		return e
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "nano"
}
