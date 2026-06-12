# Command Struct Internals

The full `Command` struct with all fields — both exported and unexported.

> **Source of truth**: <https://github.com/spf13/cobra/blob/main/command.go>. For user-facing field descriptions, see the **cobra** skill → `references/commands.md`.

## Exported Fields

### Identity and Documentation

| Field | Type | Purpose |
|-------|------|---------|
| `Use` | `string` | One-line usage message. First word = command name. |
| `Aliases` | `[]string` | Alternative names for the command |
| `SuggestFor` | `[]string` | Command names for which this command is suggested (not aliases — only suggestions) |
| `Short` | `string` | Short description shown in parent's help |
| `GroupID` | `string` | Group ID for help grouping under parent |
| `Long` | `string` | Long description shown in `help <this-command>` |
| `Example` | `string` | Example usage strings |
| `Deprecated` | `string` | Non-empty = deprecated; string is the deprecation message |
| `Annotations` | `map[string]string` | Key/value pairs for application-specific metadata |
| `Version` | `string` | Enables `--version` flag when non-empty |
| `Hidden` | `bool` | Hide from help listings (still callable) |

### Completion Fields

| Field | Type | Purpose |
|-------|------|---------|
| `ValidArgs` | `[]Completion` | Static positional arg completions (tab-separated descriptions) |
| `ValidArgsFunction` | `CompletionFunc` | Dynamic positional arg completion function |
| `ArgAliases` | `[]string` | Aliases for ValidArgs (accepted by OnlyValidArgs but not shown in completions) |
| `BashCompletionFunction` | `string` | Legacy bash V1 custom function |
| `CompletionOptions` | `CompletionOptions` | Controls auto-generated completion command behavior |

### Argument Validation

| Field | Type | Purpose |
|-------|------|---------|
| `Args` | `PositionalArgs` | Argument validator function (`nil` = `legacyArgs`) |

### Run Functions (in execution order)

| Field | Type | Inherited | Error-returning |
|-------|------|-----------|-----------------|
| `PersistentPreRun` | `func(*Command, []string)` | Yes | No |
| `PersistentPreRunE` | `func(*Command, []string) error` | Yes | Yes |
| `PreRun` | `func(*Command, []string)` | No | No |
| `PreRunE` | `func(*Command, []string) error` | No | Yes |
| `Run` | `func(*Command, []string)` | — | No |
| `RunE` | `func(*Command, []string) error` | — | Yes |
| `PostRun` | `func(*Command, []string)` | No | No |
| `PostRunE` | `func(*Command, []string) error` | No | Yes |
| `PersistentPostRun` | `func(*Command, []string)` | Yes | No |
| `PersistentPostRunE` | `func(*Command, []string) error` | Yes | Yes |

When both `*E` and non-`*E` are defined, the `*E` variant takes precedence.

### Behavior Flags

| Field | Type | Default | Purpose |
|-------|------|---------|---------|
| `TraverseChildren` | `bool` | `false` | Use `Traverse` instead of `Find` for command resolution |
| `SilenceErrors` | `bool` | `false` | Don't print "Error: ..." to stderr |
| `SilenceUsage` | `bool` | `false` | Don't print usage on error |
| `DisableFlagParsing` | `bool` | `false` | Skip flag parsing; pass raw args to Run |
| `DisableAutoGenTag` | `bool` | `false` | Skip "Auto generated" tag in man pages |
| `DisableFlagsInUseLine` | `bool` | `false` | Don't add `[flags]` to usage line |
| `DisableSuggestions` | `bool` | `false` | Disable "did you mean?" suggestions |
| `SuggestionsMinimumDistance` | `int` | `0` | Override minimum Levenshtein distance for suggestions (default 2) |

## Unexported Fields

### Flag Sets

| Field | Type | Purpose |
|-------|------|---------|
| `flags` | `*flag.FlagSet` | Full merged flag set (local + persistent + inherited) |
| `pflags` | `*flag.FlagSet` | Persistent flags declared on this command |
| `lflags` | `*flag.FlagSet` | Local flags cache (computed by `LocalFlags()`) |
| `iflags` | `*flag.FlagSet` | Inherited flags cache (computed by `InheritedFlags()`) |
| `parentsPflags` | `*flag.FlagSet` | All persistent flags from parent chain |
| `globNormFunc` | `func(*flag.FlagSet, string) flag.NormalizedName` | Global flag name normalization function |
| `flagErrorBuf` | `*bytes.Buffer` | Buffer for flag parse error output |
| `flagErrorFunc` | `func(*Command, error) error` | Custom flag error handler |

### Template and Help

| Field | Type | Purpose |
|-------|------|---------|
| `usageFunc` | `func(*Command) error` | Custom usage function |
| `usageTemplate` | `*tmplFunc` | Compiled usage template |
| `helpTemplate` | `*tmplFunc` | Compiled help template |
| `helpFunc` | `func(*Command, []string)` | Custom help function |
| `helpCommand` | `*Command` | The help subcommand |
| `helpCommandGroupID` | `string` | Group ID for the help command |
| `completionCommandGroupID` | `string` | Group ID for the completion command |
| `versionTemplate` | `*tmplFunc` | Compiled version template |
| `errPrefix` | `string` | Error prefix (default "Error:") |

### Command Tree

| Field | Type | Purpose |
|-------|------|---------|
| `commands` | `[]*Command` | Direct children |
| `parent` | `*Command` | Parent command |
| `commandgroups` | `[]*Group` | Command groups for help |
| `commandsMaxUseLen` | `int` | Max `Use` length among children (for alignment) |
| `commandsMaxCommandPathLen` | `int` | Max `CommandPath()` length among children |
| `commandsMaxNameLen` | `int` | Max `Name()` length among children |
| `commandsAreSorted` | `bool` | Whether children are sorted (controlled by `EnableCommandSorting`) |

### Execution State

| Field | Type | Purpose |
|-------|------|---------|
| `args` | `[]string` | Arguments passed to this command |
| `commandCalledAs` | `struct{name string; called bool}` | Tracks which name/alias was used to invoke this command |
| `ctx` | `context.Context` | Context (set by `ExecuteContext`/`ExecuteContextC`) |
| `inReader` | `io.Reader` | Stdin (default: `os.Stdin`) |
| `outWriter` | `io.Writer` | Stdout (default: `os.Stdout`) |
| `errWriter` | `io.Writer` | Stderr (default: `os.Stderr`) |

### Validation

| Field | Type | Purpose |
|-------|------|---------|
| `FParseErrWhitelist` | `FParseErrWhitelist` | Which flag parse errors to ignore |

```go
type FParseErrWhitelist struct {
    UnknownFlags bool  // ignore unknown flag errors
}
```

## Key Relationships

```
Command
  ├── .parent → Command (single parent)
  ├── .commands → []*Command (children)
  ├── .flags → merged FlagSet (local + persistent + inherited)
  ├── .pflags → persistent FlagSet (this command only)
  ├── .lflags → local FlagSet (computed, cached)
  ├── .iflags → inherited FlagSet (computed, cached)
  ├── .parentsPflags → all parents' persistent flags
  └── .helpCommand → the help subcommand
```

## Field Initialization

Most fields are zero-initialized. Flag sets are lazily created on first access:

```go
func (c *Command) Flags() *flag.FlagSet {
    if c.flags == nil {
        c.flags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
        // ...
    }
    return c.flags
}
```

This means `Flags()`, `PersistentFlags()`, `LocalFlags()`, and `InheritedFlags()` are all lazy — they allocate on first call.

## References

- [cobra source: command.go](https://github.com/spf13/cobra/blob/main/command.go) — struct definition (lines 54–260)

## Related Skills

- For the execute flow that uses these fields, see [references/execute-flow.md](references/execute-flow.md).
- For flag set hierarchy details, see [references/flag-resolution.md](references/flag-resolution.md).
