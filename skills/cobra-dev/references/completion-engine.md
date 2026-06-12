# Completion Engine Internals

How cobra's `__complete` hidden command and `getCompletions` work internally.

> **Source of truth**: <https://github.com/spf13/cobra/blob/main/completions.go>. For user-facing completion API, see the **cobra** skill → `references/completions.md`.

## The Hidden Commands

Cobra adds two hidden subcommands during `ExecuteC`:

| Command | Purpose |
|---------|---------|
| `__complete` | Request completions with descriptions |
| `__completeNoDesc` | Request completions without descriptions |

These are only kept in the command tree if they're actually being invoked (to avoid side effects on programs that accept arbitrary arguments).

### Constants

```go
ShellCompRequestCmd = "__complete"
ShellCompNoDescRequestCmd = "__completeNoDesc"
```

## initCompleteCmd

```go
func (c *Command) initCompleteCmd(args []string)
```

Adds the `__complete` and `__completeNoDesc` commands. Their `Run` function:

1. Calls `cmd.getCompletions(args)` to resolve completions
2. Filters Active Help entries (if disabled)
3. Strips descriptions (if `__completeNoDesc`)
4. Writes each completion to stdout (one per line)
5. Writes the directive as `:<integer>` on the last line
6. Prints directive summary to stderr (ignored by shell scripts)

## getCompletions — The Core Logic

```go
func (c *Command) getCompletions(args []string) ([]Completion, ShellCompDirective, error)
```

### Flow

```
1. Extract toComplete — the last argument (partially-typed word)
2. Find the real command:
   - Temporarily remove __complete from the command tree
   - Use Traverse() or Find() (based on TraverseChildren)
   - Restore __complete
3. checkIfFlagCompletion() — detect if completing a flag value
4. Parse flags early (to determine required flags, arg count)
5. Check for --help / --version → stop completion
6. Flag name completion (if toComplete starts with "-"):
   a. Required flags first (BashCompOneRequiredFlag annotation)
   b. All available flags (inherited + non-inherited)
   c. Skip already-set flags (unless they accept multiple values)
   d. Return ShellCompDirectiveNoFileComp (or NoSpace for single =-suffixed flag)
7. Subcommand completion (if no args and no local non-persistent flags set):
   a. List available subcommands with descriptions
   b. Return ShellCompDirectiveNoFileComp
8. ValidArgs completion (if ValidArgs is set):
   a. Return static completions from ValidArgs
9. Custom completion function:
   a. For flags: look up flagCompletionFunctions[flag]
   b. For args: call finalCmd.ValidArgsFunction
10. Return completions + directive
```

### checkIfFlagCompletion

Detects whether the user is completing a flag value:

```go
func checkIfFlagCompletion(...) (string, *pflag.Flag, error)
```

Handles these patterns:

| Pattern | Detection |
|---------|-----------|
| `--flag=value` | `strings.Contains(arg, "=")` — split on `=` |
| `--flag value` | Previous arg is a flag that expects a value |
| `-f value` | Previous arg is a shorthand that expects a value |
| `-f=value` | Shorthand with `=` delimiter |

Returns the flag name and the `*pflag.Flag` for completion lookup.

### Flag Name Completion

When `toComplete` starts with `-`:

1. **Required flags** — flags with `BashCompOneRequiredFlag` annotation are listed first
2. **Available flags** — from `cmd.Flags()` (merged set), excluding:
   - Already-set flags (unless they accept multiple values: Slice, Array, stringTo)
   - Hidden flags
   - Help and version flags (if they match the prefix)
3. **Directive** — `ShellCompDirectiveNoFileComp` (no file completion for flag names)
4. **Special case** — if exactly one flag matches and it uses `=` delimiter, return `ShellCompDirectiveNoSpace`

### Subcommand Completion

When no positional args are provided and no local non-persistent flags are set:

1. List all available subcommands (not hidden, not deprecated)
2. Include descriptions via `CompletionWithDesc`
3. Include aliases as separate entries
4. Return `ShellCompDirectiveNoFileComp`

### ValidArgs Completion

Static completions from the `ValidArgs` field:

1. Each entry in `ValidArgs` is a `Completion` (may include tab-separated description)
2. Filter by prefix matching against `toComplete`
3. Return with `ShellCompDirectiveNoFileComp`

### Custom Completion Function

For dynamic completions:

**Flag completions** — looked up from the global registry:

```go
// Global registry (mutex-protected)
var flagCompletionFunctions = map[*pflag.Flag]CompletionFunc{}
var flagCompletionMutex sync.RWMutex
```

**Argument completions** — called from `ValidArgsFunction`:

```go
comps, directive = finalCmd.ValidArgsFunction(finalCmd, finalCmd.Args, toComplete)
```

## ShellCompDirective — Bitmask

```go
type ShellCompDirective int

const (
    ShellCompDirectiveDefault       ShellCompDirective = 0
    ShellCompDirectiveError         ShellCompDirective = 1
    ShellCompDirectiveNoSpace       ShellCompDirective = 2
    ShellCompDirectiveNoFileComp    ShellCompDirective = 4
    ShellCompDirectiveFilterFileExt ShellCompDirective = 8
    ShellCompDirectiveFilterDirs    ShellCompDirective = 16
    ShellCompDirectiveKeepOrder     ShellCompDirective = 32
)
```

### Output Format

The `__complete` command writes:

```
completion1
completion2\tDescription for completion2
completion3
:<directive_integer>
```

- Each completion on its own line
- Tab-separated descriptions (if using `__complete`)
- Last line is `:` followed by the directive integer
- Directive summary printed to stderr (for debugging)

## RegisterFlagCompletionFunc — Global Registry

```go
func (c *Command) RegisterFlagCompletionFunc(flagName string, f CompletionFunc) error
```

Stores the function in a **package-level** map keyed by `*pflag.Flag` pointer:

```go
var flagCompletionFunctions = map[*pflag.Flag]CompletionFunc{}
```

Protected by `flagCompletionMutex` (sync.RWMutex). This means:
- The registry is global across all commands
- Flag pointer identity matters — the same flag name on different commands is a different entry
- Thread-safe for concurrent registration

## CompletionOptions

```go
type CompletionOptions struct {
    DisableDefaultCmd          bool
    DisableNoDescFlag          bool
    DisableDescriptions        bool
    HiddenDefaultCmd           bool
    DefaultShellCompDirective  *ShellCompDirective
}
```

| Field | Effect |
|-------|--------|
| `DisableDefaultCmd` | Don't create the `completion` subcommand |
| `DisableNoDescFlag` | Don't add `--no-descriptions` to completion subcommands |
| `DisableDescriptions` | Always strip descriptions (equivalent to always using `__completeNoDesc`) |
| `HiddenDefaultCmd` | Make the `completion` command hidden |
| `DefaultShellCompDirective` | Override the default directive when none is determined |

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `COBRA_ACTIVE_HELP` | Disable active help globally |
| `<PROG>_ACTIVE_HELP` | Disable active help per-program |
| `COBRA_COMPLETION_DESCRIPTIONS` | Runtime description control |
| `<PROG>_COMPLETION_DESCRIPTIONS` | Per-program description control |
| `BASH_COMP_DEBUG_FILE` | Debug file for bash completion |

## Edge Cases

- **__complete removal during Find**: The `__complete` command is temporarily removed from the tree during `Find()` to avoid it being matched as a subcommand. It's restored after.
- **Flag value with =**: When completing `--flag=par`, cobra detects the `=`, extracts `par` as `toComplete`, and looks up the flag's completion function.
- **Already-set flags**: Flags that have been set are excluded from flag name completion unless they're Slice/Array/stringTo types (which accept multiple values).
- **Completion command and arbitrary args**: If the root has no real subcommands, the `completion` command is only kept if it's actually being called. This prevents breaking programs that accept arbitrary arguments.
- **Active Help filtering**: Active Help entries (prefixed with `_activeHelp_ `) are filtered out if `GetActiveHelpConfig` returns `"0"`.

## References

- [cobra source: completions.go](https://github.com/spf13/cobra/blob/main/completions.go)
- [cobra source: active_help.go](https://github.com/spf13/cobra/blob/main/active_help.go)

## Related Skills

- For how shell scripts call `__complete`, see [references/shell-completions.md](references/shell-completions.md).
- For how `Find`/`Traverse` work, see [references/execute-flow.md](references/execute-flow.md).
