# exe

A personal command launcher. `exe <app> <command>` resolves to a shell command
and runs it in your current terminal, driven by a plain config file you edit
yourself, not by anything hardcoded in this repo.

## Install

```powershell
# Windows
irm https://raw.githubusercontent.com/prjvvl/exe/main/install.ps1 | iex
```

```bash
# Mac/Linux
curl -fsSL https://raw.githubusercontent.com/prjvvl/exe/main/install.sh | sh
```

The installer also adds the shell hook to your profile automatically (this
is what lets exe actually run commands in your terminal instead of just
printing them), so all you need after that is to restart your terminal.

## Usage

```
exe                    # interactive menu
exe nav                # menu scoped to the "nav" app
exe nav projects       # runs the "projects" command directly
```

exe never executes anything itself. It resolves a command to a shell string
and hands it to the shell hook to run in your actual terminal, that's what
lets `exe nav projects` actually `cd` your shell.

## Config

Lives at `os.UserConfigDir()/exe/config.toml`, empty by default.

```toml
[[apps]]
name = "nav"
description = "Jump to common folders"

  [[apps.commands]]
  name = "projects"
  description = "Go to your projects folder"
  run = "cd ~/projects"
```

`edit`, `view`, `info`, `init`, `update` are reserved names. The file is
validated on every load, so a malformed edit (by you, or an AI agent editing
it directly) fails with a specific error instead of misbehaving.

## Built-in commands

- `exe info`: explains exe and shows the current config (the entry point for
  pointing an AI agent at it)
- `exe view`: prints the raw config file
- `exe edit`: opens the config file in your editor
- `exe update`: checks for and installs the latest release
- `exe init <shell>`: prints the shell-hook snippet (`powershell`, `bash`, `zsh`)

## Landing page

**https://prjvvl.github.io/exe/**

Static install page, served from `docs/index.html` via GitHub Pages
(Settings > Pages > Deploy from a branch > `main` / `/docs`).

## Building from source

```powershell
go build -o exe.exe .
```

Put the binary on your `PATH`, then install the shell hook manually (the
install scripts do this automatically, but a source build doesn't go
through them):

```powershell
exe init powershell | Out-String | Invoke-Expression   # add to $PROFILE
```

```bash
eval "$(exe init bash)"   # add to .bashrc, or use "zsh" + .zshrc
```

## Status

Scaffold stage. The TUI is a plain stdin/stdout placeholder. The release
pipeline hasn't been exercised yet, no repo pushed or tag cut.

## License

[MIT](LICENSE)
