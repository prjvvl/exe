package builtin

import (
	"os"

	"github.com/prjvvl/exe/internal/config"
	"github.com/prjvvl/exe/internal/dispatch"
)

func ViewBuiltin() Builtin {
	return Builtin{
		Name:        "view",
		Description: "Print the raw config file",
		Handler: func(cfg *config.Config, args []string) (dispatch.Resolved, error) {
			path, err := config.Path()
			if err != nil {
				return dispatch.Resolved{}, err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return dispatch.Resolved{}, err
			}
			return dispatch.Resolved{Kind: dispatch.KindPrint, Text: string(data)}, nil
		},
	}
}
