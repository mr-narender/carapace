# Gen

Calling [`Gen`](https://pkg.go.dev/github.com/carapace-sh/carapace#Gen) on the root command is sufficient to enable completion script generation using the [Hidden Subcommand](#hidden-subcommand).

```go
import (
    "github.com/carapace-sh/carapace"
)

carapace.Gen(rootCmd)
```

Additionally invoke [`carapace.Test`](https://pkg.go.dev/github.com/carapace-sh/carapace#Test) in a [test](https://golang.org/doc/tutorial/add-a-test) to verify configuration during build time.
```go
func TestCarapace(t *testing.T) {
    carapace.Test(t)
}
```

## Options

`Gen` accepts [Option](https://pkg.go.dev/github.com/carapace-sh/carapace#Option) values to configure advanced behavior. Options are applied to the storage entry keyed by the command, so successive `Gen()` calls on the same command accumulate.

### WithSubcommands

[`WithSubcommands`](https://pkg.go.dev/github.com/carapace-sh/carapace#WithSubcommands) enables [multi-completer](./multiCompleter.md) support. The listed commands serve as independent completers within one binary. The binary name is automatically included as a pseudo-subcommand for self-completion.

```go
carapace.Gen(rootCmd, carapace.WithSubcommands(identifyCmd, convertCmd))
```

### WithDefault

[`WithDefault`](https://pkg.go.dev/github.com/carapace-sh/carapace#WithDefault) sets the default subcommand for multi-completer routing when `os.Args[4]` is not a known subcommand name. Defaults to the first subcommand. No-op without `WithSubcommands`.

```go
carapace.Gen(rootCmd,
    carapace.WithSubcommands(identifyCmd, convertCmd),
    carapace.WithDefault("convert"),
)
```

### WithSnippetFuncs

[`WithSnippetFuncs`](https://pkg.go.dev/github.com/carapace-sh/carapace#WithSnippetFuncs) adds custom shell code to the generated snippets. The map key is the shell name; the value is the code to inject. Multiple calls accumulate per shell in order.

```go
carapace.Gen(rootCmd, carapace.WithSnippetFuncs(map[string]string{
    "bash": "# custom bash setup\n",
    "zsh":  "# custom zsh setup\n",
}))
```

### Execute

[`Execute`](https://pkg.go.dev/github.com/carapace-sh/carapace#Carapace.Execute) intercepts `os.Args` for multi-completer routing and then calls `cmd.Execute()`. Use this instead of `cmd.Execute()`. For single completers (no `WithSubcommands`), this just calls `cmd.Execute()`.

```go
func main() {
    _ = carapace.Gen(rootCmd).Execute()
}
```

## Hidden Subcommand

When [`Gen`](https://pkg.go.dev/github.com/carapace-sh/carapace#Gen) is invoked a hidden subcommand (`_carapace`) is added. This handles completion script generation and [callbacks](./defaultActions/actionCallback.md).


### Completion

`SHELL` is optional and will be detected by parent process name.

```sh
command _carapace [SHELL]
```

```sh
# bash
source <(command _carapace)

# cmd (~/AppData/Local/clink/{command}.lua
load(io.popen('command _carapace cmd-clink'):read("*a"))()

# elvish
eval (command _carapace | slurp)

# fish
command _carapace | source

# nushell (update config.nu according to output)
command _carapace nushell

# oil
source <(command _carapace)

# powershell
Set-PSReadLineOption -Colors @{ "Selection" = "`e[7m" }
Set-PSReadlineKeyHandler -Key Tab -Function MenuComplete
command _carapace | Out-String | Invoke-Expression

# tcsh
set autolist
eval `command _carapace tcsh`

# xonsh
COMPLETIONS_CONFIRM=True
exec($(command _carapace))

# zsh
source <(command _carapace)
```

> Directly sourcing multiple completions in your shell init script increases startup time considerably. See [lazycomplete](https://github.com/rsteube/lazycomplete) for a solution to this problem.
