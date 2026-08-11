package registry

import (
	"github.com/prjvvl/exe/internal/builtin"
	"github.com/prjvvl/exe/internal/config"
)

type Registry struct {
	Config   *config.Config
	Builtins []builtin.Builtin
}

func Build(cfg *config.Config) Registry {
	return Registry{Config: cfg, Builtins: builtin.All()}
}

func (r Registry) FindBuiltin(name string) (builtin.Builtin, bool) {
	for _, b := range r.Builtins {
		if b.Name == name {
			return b, true
		}
	}
	return builtin.Builtin{}, false
}
