# Suggestions and Typo Correction

How cobra suggests similar commands when the user makes a typo.

> **Source of truth**: <https://github.com/spf13/cobra/blob/main/command.go> and <https://github.com/spf13/cobra/blob/main/cobra.go>. For user-facing suggestion configuration, see the **cobra** skill → `references/commands.md`.

## The Suggestion Flow

When `Find` or `Traverse` fails to match a subcommand, cobra checks for suggestions:

```
1. findNext(nextSubCmd) returns nil
2. In Find(): check if suggestions should be shown
3. If not disabled → compute Levenshtein distance to all subcommands
4. Suggest commands with distance ≤ SuggestionsMinimumDistance (default: 2)
5. Also check SuggestFor for explicit suggestion mappings
6. Print "Did you mean this?" message
```

## Levenshtein Distance

```go
func ld(s, t string, ignoreCase bool) int
```

Cobra implements its own Levenshtein distance algorithm (dynamic programming). The function computes the minimum number of single-character edits (insertions, deletions, substitutions) to transform `s` into `t`.

When `ignoreCase` is `true`, the comparison is case-insensitive (controlled by `EnableCaseInsensitive`).

### Default Threshold

The default `SuggestionsMinimumDistance` is `0`, which cobra internally treats as `2`:

```go
if c.SuggestionsMinimumDistance == 0 {
    c.SuggestionsMinimumDistance = 2
}
```

Only commands with a Levenshtein distance ≤ 2 are suggested.

### Example

```
$ hugo srever
Error: unknown command "srever" for "hugo"

Did you mean this?
        server
```

`"srever"` → `"server"` has Levenshtein distance 2 (swap 'e' and 'r'), which is within the threshold.

## findNext — The Matching Algorithm

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
    // 3. Single unambiguous prefix match
    if len(matches) == 1 {
        return matches[0]
    }
    return nil
}
```

### Matching Priority

1. **Exact name match** — `commandNameMatches(cmd.Name(), next)`
2. **Exact alias match** — `cmd.HasAlias(next)`
3. **Prefix match** (if `EnablePrefixMatching`) — single unambiguous prefix
4. **No match** — returns `nil`, which triggers the suggestion system

### commandNameMatches

```go
func commandNameMatches(s string, t string) bool {
    if EnableCaseInsensitive {
        return strings.EqualFold(s, t)
    }
    return s == t
}
```

### hasNameOrAliasPrefix

```go
func (c *Command) hasNameOrAliasPrefix(prefix string) bool {
    if strings.HasPrefix(c.Name(), prefix) {
        c.commandCalledAs.name = c.Name()
        return true
    }
    for _, alias := range c.Aliases {
        if strings.HasPrefix(alias, prefix) {
            c.commandCalledAs.name = alias
            return true
        }
    }
    return false
}
```

Records `commandCalledAs.name` so `CalledAs()` returns the matched name/alias.

## SuggestFor — Explicit Suggestions

The `SuggestFor` field provides explicit suggestion mappings that bypass Levenshtein distance:

```go
removeCmd := &cobra.Command{
    Use:       "rm",
    SuggestFor: []string{"remove", "delete"},
}
```

When the user types `app remove` and there's no `remove` command, cobra suggests `rm` because `"remove"` is in `rm.SuggestFor`.

## Configuration

| Field/Variable | Type | Default | Purpose |
|----------------|------|---------|---------|
| `DisableSuggestions` | `bool` | `false` | Disable the suggestion system entirely |
| `SuggestionsMinimumDistance` | `int` | `0` (treated as 2) | Maximum Levenshtein distance for suggestions |
| `SuggestFor` | `[]string` | `nil` | Explicit suggestion targets |
| `EnablePrefixMatching` | `bool` | `false` | Allow prefix matching (separate from suggestions) |
| `EnableCaseInsensitive` | `bool` | `false` | Case-insensitive matching (affects both exact and Levenshtein) |

## How Suggestions Are Displayed

When `Find` returns an error (unknown command), cobra's error handling in `ExecuteC`:

1. Checks `!c.SilenceErrors` → prints error prefix + "unknown command" message
2. If suggestions are found → prints "Did you mean this?" with matching commands
3. Prints "Run 'app --help' for usage."

## Edge Cases

- **Ambiguous prefix matches**: If two commands share a prefix, `findNext` returns `nil` (no match). The suggestion system then suggests both via Levenshtein distance.
- **EnablePrefixMatching + suggestions**: These are independent systems. Prefix matching resolves commands without going through the suggestion path. Suggestions only trigger when `findNext` returns `nil`.
- **Case-insensitive Levenshtein**: When `EnableCaseInsensitive` is true, the `ld` function compares strings case-insensitively. This means `"SRVER"` would match `"server"` with distance 2.
- **SuggestionsMinimumDistance = 1**: Only very close typos are suggested. Useful for large command sets where distance 2 produces too many false positives.
- **SuggestFor is on the target command**: `SuggestFor` is set on the command that should be suggested, not on the typo. This is the opposite of what you might expect.

## References

- [cobra source: command.go](https://github.com/spf13/cobra/blob/main/command.go) — `findNext`, `hasNameOrAliasPrefix`, `commandNameMatches`
- [cobra source: cobra.go](https://github.com/spf13/cobra/blob/main/cobra.go) — `ld`, `EnablePrefixMatching`, `EnableCaseInsensitive`

## Related Skills

- For how `findNext` fits into `Find`/`Traverse`, see [references/execute-flow.md](references/execute-flow.md).
- For global state that controls matching, see [references/global-state.md](references/global-state.md).
