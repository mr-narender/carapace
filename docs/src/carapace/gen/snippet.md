# Snippet

[`Snippet`](https://pkg.go.dev/github.com/carapace-sh/carapace#Carapace.Snippet) creates a completion script for the given shell.

For single completers, produces a standard single-command snippet. For [multi-completers](../multiCompleter.md), produces a multi-completer snippet that registers all subcommand completers at once.

```go
snippet, err := carapace.Gen(rootCmd).Snippet("bash")
```

Custom shell code can be injected via [`WithSnippetFuncs`](../gen.md#withsnippetfuncs).
