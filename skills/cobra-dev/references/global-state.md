# Global State and Package-Level Functions

Cobra's package-level variables, initialization system, and template functions.

> **Source of truth**: <https://github.com/spf13/cobra/blob/main/cobra.go>. For user-facing hook usage, see the **cobra** skill → `references/hooks.md`.

## Global Variables

### Command Resolution

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `EnablePrefixMatching` | `bool` | `false` | Allow prefix matching for subcommand names (dangerous — ambiguous prefixes return nil) |
| `EnableCommandSorting` | `bool` | `true` | Sort subcommands alphabetically in help output |
| `EnableCaseInsensitive` | `bool` | `false` | Case-insensitive command name matching |
| `EnableTraverseRunHooks` | `bool` | `false` | Execute all parents' persistent hooks (not just the first found) |

### Windows Mousetrap

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `MousetrapHelpText` | `string` | `"This is a command line tool.\n\nYou need to open cmd.exe and run it from there.\n"` | Message shown when launched from Windows Explorer |
| `MousetrapDisplayDuration` | `time.Duration` | `5 * time.Second` | How long the mousetrap message is displayed (0 = wait for key press) |

### Default Constants

```go
const (
    defaultPrefixMatching   = false
    defaultCommandSorting   = true
    defaultCaseInsensitive  = false
    defaultTraverseRunHooks = false
)
```

## OnInitialize / OnFinalize

```go
var initializers []func()
var finalizers []func()

func OnInitialize(y ...func())  // append to initializers
func OnFinalize(y ...func())    // append to finalizers
```

### When They Run

- **initializers**: Run at the start of `ExecuteC()`, before command resolution
- **finalizers**: Run at the end of `ExecuteC()`, after command execution (success or error)

### Usage Pattern

```go
cobra.OnInitialize(initConfig, initLogging)
cobra.OnFinalize(cleanupResources)
```

### Key Details

- **Append-only**: Each call appends to the slice. There's no way to remove or replace.
- **Run on every execution**: Including completions, help, and version commands.
- **No `init()` function**: Cobra doesn't use Go's `init()`. All initialization is explicit via `OnInitialize`.
- **Ordering**: Functions run in the order they were registered.

## Template Functions

```go
var templateFuncs = template.FuncMap{
    "trim":                    strings.TrimSpace,
    "trimRightSpace":          trimRightSpace,
    "trimTrailingWhitespaces": trimRightSpace,
    "appendIfNotPresent":      appendIfNotPresent,
    "rpad":                    rpad,
    "gt":                      Gt,
    "eq":                      Eq,
}
```

### AddTemplateFunc / AddTemplateFuncs

```go
func AddTemplateFunc(name string, tmplFunc interface{})
func AddTemplateFuncs(tmplFuncs template.FuncMap)
```

Add custom functions available in help, usage, and version templates.

### Built-in Template Functions

| Function | Signature | Purpose |
|----------|-----------|---------|
| `trim` | `func(string) string` | `strings.TrimSpace` |
| `trimRightSpace` | `func(string) string` | Strip trailing whitespace (unicode.IsSpace) |
| `trimTrailingWhitespaces` | `func(string) string` | Alias for `trimRightSpace` |
| `appendIfNotPresent` | `func(s, append string) string` | Append if not already present |
| `rpad` | `func(string, int) string` | Right-pad to width |
| `gt` | `func(a, b interface{}) bool` | Greater-than comparison |
| `eq` | `func(a, b interface{}) bool` | Equality comparison |

### Deprecated Functions

`Gt` and `Eq` are marked with `FIXME` comments — they are unused by cobra itself and should be removed in v2. They exist only for backward compatibility with user templates.

## tmpl — Template Compilation

```go
type tmplFunc struct {
    template *template.Template
}

func tmpl(text string) *tmplFunc {
    t := template.New("").Funcs(templateFuncs)
    template.Must(t.Parse(text))
    return &tmplFunc{template: t}
}
```

Templates are compiled once (when `SetHelpTemplate`, `SetUsageTemplate`, or `SetVersionTemplate` is called) and cached in the `Command` struct's `helpTemplate`, `usageTemplate`, and `versionTemplate` fields.

## CheckErr

```go
func CheckErr(msg interface{}) {
    if msg != nil {
        fmt.Fprintln(os.Stderr, "Error:", msg)
        os.Exit(1)
    }
}
```

A convenience function for simple error handling. Prints "Error:" prefix and exits with code 1.

## WriteStringAndCheck

```go
func WriteStringAndCheck(b io.StringWriter, s string) {
    _, err := b.WriteString(s)
    CheckErr(err)
}
```

Used internally for completion script generation. Panics on write errors.

## Edge Cases

- **EnablePrefixMatching is dangerous**: If two subcommands share a prefix (e.g., `serve` and `server`), typing `ser` is ambiguous and returns nil. Only use this when subcommand names are clearly distinct.
- **EnableCommandSorting affects help only**: Subcommands are sorted in help output but not in the `commands` slice itself. The `commandsAreSorted` flag tracks whether sorting has been applied.
- **OnInitialize runs before everything**: Including before `Find`/`Traverse`. This means initializers can modify the command tree (add commands, change flags) before resolution.
- **Mousetrap is Windows-only**: The `preExecHookFn` is only set on Windows (in `command_win.go`). On other platforms, it's nil.
- **templateFuncs is global**: Adding a template function affects all commands in the process. There's no per-command template function scope.

## References

- [cobra source: cobra.go](https://github.com/spf13/cobra/blob/main/cobra.go)
- [cobra source: command_win.go](https://github.com/spf13/cobra/blob/main/command_win.go) — mousetrap hook
- [cobra source: command_notwin.go](https://github.com/spf13/cobra/blob/main/command_notwin.go) — no-op hook

## Related Skills

- For how `OnInitialize` is consumed in `ExecuteC`, see [references/execute-flow.md](references/execute-flow.md).
- For how `EnablePrefixMatching` affects `findNext`, see [references/suggestions.md](references/suggestions.md).
