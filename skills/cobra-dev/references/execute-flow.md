# Execute Flow: ExecuteC, Find, Traverse

How cobra resolves which command to run and executes it.

> **Source of truth**: <https://github.com/spf13/cobra/blob/main/command.go>. For user-facing hook behavior, see the **cobra** skill → `references/hooks.md`.

## ExecuteC — The Main Entry Point

```go
func (c *Command) ExecuteC() (cmd *Command, err error)
```

### Flow

```
1. Set context (default: context.Background())
2. Redirect to root: if c.HasParent() → c.Root().ExecuteC()
3. Windows mousetrap hook (preExecHookFn)
4. InitDefaultHelpCmd() — add help command if subcommands exist
5. Get args (default: os.Args[1:])
6. initCompleteCmd(args) — add __complete hidden command
7. InitDefaultCompletionCmd(args) — add completion subcommand
8. checkCommandGroups() — validate group IDs
9. Command resolution:
   if TraverseChildren → Traverse(args)
   else                → Find(args)
10. Set commandCalledAs
11. Propagate context to resolved command
12. cmd.execute(flags) — run the command
13. Handle errors (flag.ErrHelp → show help; others → print + return)
```

### Key Details

- **Always starts at root**: `ExecuteC` on any subcommand redirects to the root. This ensures `OnInitialize` and help setup run correctly.
- **Mousetrap**: On Windows, if the program is launched from Explorer (not cmd.exe), cobra shows `MousetrapHelpText` and waits. Controlled by `MousetrapHelpText` and `MousetrapDisplayDuration` global variables.
- **Args source**: `c.args` takes precedence, then `os.Args[1:]`. Exception: `cobra.test` binary (workaround for Go test -v).

## Find — Default Resolution

```go
func (c *Command) Find(args []string) (*Command, []string, error)
```

### Algorithm

```
1. stripFlags(innerArgs, c) — remove all flag-like args and their values
2. If no positional args remain → return current command
3. Take first positional arg as nextSubCmd
4. findNext(nextSubCmd) — look up subcommand
5. If found → recurse on matched subcommand with argsMinusFirstX(innerArgs, nextSubCmd)
6. If not found → return current command with original args
7. Apply legacyArgs if commandFound.Args == nil
```

### stripFlags

Removes flags and their values from the args slice:

```go
func stripFlags(args []string, c *Command) []string
```

- Skips `--` (treats everything after as positional)
- Removes `--flag value` pairs (checks if flag expects a value)
- Removes `--flag=value` (single arg)
- Removes `-f value` pairs
- Removes `-fvalue` combined args
- Keeps everything else as positional

### argsMinusFirstX

Removes only the **first occurrence** of `x` from args, being careful not to remove flag values:

```go
// Handles: openshift admin policy add-role-to-user admin my-user
// "admin" appears as both subcommand and positional arg
```

This is important when the same word appears as both a subcommand name and a flag value or positional arg.

### legacyArgs

Applied when `cmd.Args == nil`:

| Condition | Behavior |
|-----------|----------|
| Root with subcommands + remaining args | Return `"unknown command"` error |
| Root without subcommands | Accept arbitrary args |
| Subcommand | Accept arbitrary args |

This is why root commands with subcommands reject unknown positional args by default.

## Traverse — Alternative Resolution

```go
func (c *Command) Traverse(args []string) (*Command, []string, error)
```

Enabled by `TraverseChildren = true`.

### Algorithm

```
1. Iterate args one by one
2. Accumulate flags:
   - --flag (no =) → check if expects value, set inFlag
   - -f (len 2, no =) → check if expects value, set inFlag
   - If inFlag → consume as flag value
   - --flag=value or -fvalue → consume as flag
3. On first non-flag arg:
   - findNext(arg) → look up subcommand
   - If found → ParseFlags(accumulated flags) on current command, then recurse
   - If not found → return current command with all remaining args
4. If all args are flags → return current command
```

### Find vs Traverse Comparison

| Aspect | Find | Traverse |
|--------|------|----------|
| Flag handling | Strips all flags upfront | Accumulates and parses flags per level |
| Flag parsing | Does NOT parse flags on parents | Parses flags on each parent via `ParseFlags` |
| Use case | Simple, fast resolution | Needed when parent flags must be processed |
| Activation | Default (`TraverseChildren = false`) | `TraverseChildren = true` |
| Matching | Same `findNext` | Same `findNext` |
| Performance | Faster (no flag parsing during traversal) | Slower (parses flags at each level) |

## findNext — The Core Matcher

Used by both `Find` and `Traverse`:

```go
func (c *Command) findNext(next string) *Command {
    matches := make([]*Command, 0)
    for _, cmd := range c.commands {
        // 1. Exact name match
        if commandNameMatches(cmd.Name(), next) || cmd.HasAlias(next) {
            cmd.commandCalledAs.name = next
            return cmd
        }
        // 2. Prefix match (only if EnablePrefixMatching)
        if EnablePrefixMatching && cmd.hasNameOrAliasPrefix(next) {
            matches = append(matches, cmd)
        }
    }
    // 3. Return single prefix match; nil if ambiguous or no match
    if len(matches) == 1 {
        return matches[0]
    }
    return nil
}
```

### Priority

1. **Exact name match** — `commandNameMatches(cmd.Name(), next)`
2. **Exact alias match** — `cmd.HasAlias(next)`
3. **Prefix match** (if `EnablePrefixMatching`) — single unambiguous prefix
4. **No match** — returns `nil`

### commandNameMatches

```go
func commandNameMatches(s string, t string) bool {
    if EnableCaseInsensitive {
        return strings.EqualFold(s, t)
    }
    return s == t
}
```

Case-insensitive matching is controlled by the global `EnableCaseInsensitive` variable.

## execute — The Private Runner

```go
func (c *Command) execute(a []string) (err error)
```

### Full Execution Sequence

```
1. Nil check: panic if c == nil
2. Print deprecation warning (if Deprecated is set)
3. InitDefaultHelpFlag() — add --help if not present
4. InitDefaultVersionFlag() — add --version if Version is set
5. ParseFlags(a) — parse flags into c.flags
6. Flag error handling: call FlagErrorFunc (default: print error + usage)
7. Check --help → return flag.ErrHelp
8. Check --version → render version template and return
9. Check Runnable() → return flag.ErrHelp if no Run/RunE
10. preRun() — execute persistent pre-run chain + pre-run
11. ValidateArgs(argWoFlags) — validate positional args
12. ValidateRequiredFlags() — check required flags
13. ValidateFlagGroups() — check flag group constraints
14. Run / RunE — main command logic
15. PostRun / PostRunE
16. postRun() — execute persistent post-run chain
```

### preRun / postRun

```go
func (c *Command) preRun() {
    // Build parent chain (root → ... → current)
    parents := make([]*Command, 0, 5)
    for p := c; p != nil; p = p.Parent() {
        parents = append(parents, p)
    }
    // Reverse: root first, current last
    for i := len(parents) - 1; i >= 0; i-- {
        p := parents[i]
        if EnableTraverseRunHooks {
            // Always run all parents' persistent hooks
        } else {
            // Only run the first persistent hook found (skip rest)
        }
    }
    // Then run current command's PreRun
}
```

The `postRun` reverses the order: current command's `PostRun` first, then parents' `PersistentPostRun` from innermost to outermost.

## Edge Cases

- **Find strips all flags**: This means flags on parent commands are not parsed during `Find`. They're parsed later in `execute()`. This is why `TraverseChildren` exists — it parses flags at each level during traversal.
- **Ambiguous prefix matches return nil**: If two subcommands share a prefix, neither is selected. The user must type enough to disambiguate.
- **commandCalledAs tracking**: `findNext` records which name/alias was used. `CalledAs()` returns this after execution.
- **legacyArgs is a footgun**: Root commands with subcommands reject unknown args. This surprises developers who expect `ArbitraryArgs` behavior. Set `Args: cobra.ArbitraryArgs` to override.

## References

- [cobra source: command.go](https://github.com/spf13/cobra/blob/main/command.go) — `ExecuteC`, `Find`, `Traverse`, `execute`, `findNext`, `stripFlags`, `argsMinusFirstX`

## Related Skills

- For flag set hierarchy and `ParseFlags`, see [references/flag-resolution.md](references/flag-resolution.md).
- For the completion engine's use of `Find`/`Traverse`, see [references/completion-engine.md](references/completion-engine.md).
