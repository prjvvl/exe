package shellhook

import "fmt"

func Generate(shell string) (string, error) {
	switch shell {
	case "powershell", "pwsh":
		return powershellHook, nil
	case "bash":
		return bashHook, nil
	case "zsh":
		return zshHook, nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: powershell, bash, zsh)", shell)
	}
}

const powershellHook = `function exe {
    $exeBin = (Get-Command -CommandType Application exe -ErrorAction SilentlyContinue | Select-Object -First 1).Source
    if (-not $exeBin) {
        Write-Error "exe: could not find the exe binary on PATH"
        return
    }

    $evalFile = [System.IO.Path]::GetTempFileName()
    Remove-Item $evalFile -ErrorAction SilentlyContinue

    try {
        $env:EXE_EVAL_FILE = $evalFile
        & $exeBin @args
    } finally {
        Remove-Item Env:\EXE_EVAL_FILE -ErrorAction SilentlyContinue
    }

    if (Test-Path $evalFile) {
        $toRun = Get-Content -Raw $evalFile -ErrorAction SilentlyContinue
        Remove-Item $evalFile -ErrorAction SilentlyContinue
        if ($toRun -and $toRun.Trim()) {
            Invoke-Expression $toRun
        }
    }
}`

const bashHook = `exe() {
    local exe_bin
    exe_bin="$(type -P exe)"
    if [ -z "$exe_bin" ]; then
        echo "exe: could not find the exe binary on PATH" >&2
        return 1
    fi

    local eval_file
    eval_file="$(mktemp)"
    rm -f "$eval_file"

    EXE_EVAL_FILE="$eval_file" "$exe_bin" "$@"

    if [ -f "$eval_file" ]; then
        local to_run
        to_run="$(cat "$eval_file")"
        rm -f "$eval_file"
        if [ -n "$to_run" ]; then
            eval "$to_run"
        fi
    fi
}`

// zsh's builtin lookup semantics differ from bash's (command -v / type would
// also match a shell function), so it uses `whence -p` to search PATH only.
const zshHook = `exe() {
    local exe_bin
    exe_bin="$(whence -p exe)"
    if [ -z "$exe_bin" ]; then
        echo "exe: could not find the exe binary on PATH" >&2
        return 1
    fi

    local eval_file
    eval_file="$(mktemp)"
    rm -f "$eval_file"

    EXE_EVAL_FILE="$eval_file" "$exe_bin" "$@"

    if [ -f "$eval_file" ]; then
        local to_run
        to_run="$(cat "$eval_file")"
        rm -f "$eval_file"
        if [ -n "$to_run" ]; then
            eval "$to_run"
        fi
    fi
}`
