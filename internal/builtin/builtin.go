package builtin

import (
	"github.com/prjvvl/exe/internal/config"
	"github.com/prjvvl/exe/internal/dispatch"
)

type Handler func(cfg *config.Config, args []string) (dispatch.Resolved, error)

type Builtin struct {
	Name        string
	Description string
	Handler     Handler
}

func All() []Builtin {
	return []Builtin{
		InfoBuiltin(),
		ViewBuiltin(),
		EditBuiltin(),
		UpdateBuiltin(),
	}
}
