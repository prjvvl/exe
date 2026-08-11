package dispatch

import (
	"fmt"
	"os"
)

type Kind int

const (
	KindPrint Kind = iota
	KindEval
)

type Resolved struct {
	Kind Kind
	Text string
}

const EvalFileEnv = "EXE_EVAL_FILE"

func Emit(r Resolved) error {
	if r.Kind != KindEval {
		fmt.Println(r.Text)
		return nil
	}

	if evalFile := os.Getenv(EvalFileEnv); evalFile != "" {
		return os.WriteFile(evalFile, []byte(r.Text), 0o600)
	}

	fmt.Println(r.Text)
	fmt.Fprintln(os.Stderr, "exe: no shell hook detected (run `exe init <shell>` and add it to your profile), printed instead of running")
	return nil
}
