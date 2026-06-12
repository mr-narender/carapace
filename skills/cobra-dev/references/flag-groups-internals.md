# Flag Group Validation Internals

How cobra implements flag groups using pflag annotations.

> **Source of truth**: <https://github.com/spf13/cobra/blob/main/flag_groups.go>. For user-facing flag group usage, see the **cobra** skill → `references/flags.md`.

## Annotation Constants

Flag groups are implemented entirely through pflag annotations:

```go
const (
    requiredAsGroupAnnotation   = "cobra_annotation_required_if_others_set"
    oneRequiredAnnotation       = "cobra_annotation_one_required"
    mutuallyExclusiveAnnotation = "cobra_annotation_mutually_exclusive"
)
```

## MarkFlagsRequiredTogether

```go
func (c *Command) MarkFlagsRequiredTogether(flagNames ...string) {
    c.mergePersistentFlags()
    for _, v := range flagNames {
        f := c.Flags().Lookup(v)
        if f == nil {
            panic(fmt.Sprintf("Failed to find flag %q and mark it as being required in a flag group", v))
        }
        if err := c.Flags().SetAnnotation(v, requiredAsGroupAnnotation,
            append(f.Annotations[requiredAsGroupAnnotation], strings.Join(flagNames, " ")),
        ); err != nil {
            panic(err)
        }
    }
}
```

### How It Works

1. Calls `mergePersistentFlags()` to ensure all flags are available
2. For each flag in the group:
   - Looks up the flag in the merged `Flags()` set
   - Panics if the flag doesn't exist
   - Appends the space-joined group member names to the flag's annotation

### Annotation Storage

Each flag can have **multiple** annotation entries for the same key. This allows a flag to belong to multiple groups:

```
flag "username":
  cobra_annotation_required_if_others_set: ["username password", "username token"]

flag "password":
  cobra_annotation_required_if_others_set: ["username password"]

flag "token":
  cobra_annotation_required_if_others_set: ["username token"]
```

The annotation value is a space-joined string of all flags in the group. This is parsed back during validation.

## MarkFlagsOneRequired and MarkFlagsMutuallyExclusive

Same structure, different annotation key:

```go
MarkFlagsOneRequired → oneRequiredAnnotation = "cobra_annotation_one_required"
MarkFlagsMutuallyExclusive → mutuallyExclusiveAnnotation = "cobra_annotation_mutually_exclusive"
```

## ValidateFlagGroups — The Validation Entry Point

```go
func (c *Command) ValidateFlagGroups() error {
    if c.DisableFlagParsing {
        return nil
    }

    flags := c.Flags()

    groupStatus := map[string]map[string]bool{}
    oneRequiredGroupStatus := map[string]map[string]bool{}
    mutuallyExclusiveGroupStatus := map[string]map[string]bool{}

    flags.VisitAll(func(pflag *flag.Flag) {
        processFlagForGroupAnnotation(flags, pflag, requiredAsGroupAnnotation, groupStatus)
        processFlagForGroupAnnotation(flags, pflag, oneRequiredAnnotation, oneRequiredGroupStatus)
        processFlagForGroupAnnotation(flags, pflag, mutuallyExclusiveAnnotation, mutuallyExclusiveGroupStatus)
    })

    if err := validateRequiredFlagGroups(groupStatus); err != nil {
        return err
    }
    if err := validateOneRequiredFlagGroups(oneRequiredGroupStatus); err != nil {
        return err
    }
    if err := validateExclusiveFlagGroups(mutuallyExclusiveGroupStatus); err != nil {
        return err
    }
    return nil
}
```

### Validation Order

1. **Required-together** groups first
2. **One-required** groups second
3. **Mutually-exclusive** groups last

Returns the **first** error encountered. Short-circuits on error.

## processFlagForGroupAnnotation

```go
func processFlagForGroupAnnotation(
    flags *flag.FlagSet,
    pflag *flag.Flag,
    annotation string,
    groupStatus map[string]map[string]bool,
)
```

### Algorithm

1. Check if the flag has the given annotation
2. For each annotation value (space-joined group member names):
   a. Split the value into individual flag names
   b. Check if **all** flags in the group exist on this command (`hasAllFlags`)
   c. If not all exist → skip this group (not enforced here)
   d. If all exist → create a status map: `{flagName → Changed}`
3. Record `pflag.Changed` (whether the flag was explicitly set by the user)

### The hasAllFlags Check

```go
func hasAllFlags(fs *flag.FlagSet, flagnames ...string) bool {
    for _, fname := range flagnames {
        if fs.Lookup(fname) == nil {
            return false
        }
    }
    return true
}
```

This is why "a group is only enforced on commands where every flag in the group is defined". If a parent defines `--username` and `--password` as a required-together group, but a child only inherits `--username`, the group is not enforced on the child.

## Validation Rules

### validateRequiredFlagGroups

```go
func validateRequiredFlagGroups(data map[string]map[string]bool) error
```

| Some set | All set | None set | Result |
|----------|---------|----------|--------|
| ✓ | ✗ | ✗ | **Error** — missing flags listed |
| ✗ | ✓ | ✗ | ✓ — all set |
| ✗ | ✗ | ✓ | ✓ — none set (all-or-nothing) |

Error message: `if any flags in the group [username password] are set they must all be set; missing [password]`

### validateOneRequiredFlagGroups

```go
func validateOneRequiredFlagGroups(data map[string]map[string]bool) error
```

| ≥1 set | 0 set | Result |
|--------|-------|--------|
| ✓ | ✗ | ✓ |
| ✗ | ✓ | **Error** — at least one required |

Error message: `at least one of the flags in the group [json yaml] is required`

### validateExclusiveFlagGroups

```go
func validateExclusiveFlagGroups(data map[string]map[string]bool) error
```

| 0 set | 1 set | ≥2 set | Result |
|-------|--------|--------|--------|
| ✓ | ✓ | ✗ | ✓ |
| ✗ | ✗ | ✓ | **Error** — mutually exclusive |

Error message: `if any flags in the group [json yaml] are set none of the others can be; [json yaml] were all set`

## enforceFlagGroupsForCompletion

A separate function that adjusts flag behavior **during shell completion** (not during command execution):

| Group type | Completion behavior |
|------------|-------------------|
| Required-together | If one flag in group is set → mark all others as required (suggest them) |
| One-required | If none are set → mark all as required (suggest them) |
| Mutually-exclusive | If one is set → hide the others (don't suggest them) |

This makes completion "smart" about flag groups — it guides the user toward valid combinations.

## Edge Cases

- **Multiple groups per flag**: A flag can belong to multiple groups. Each group is validated independently.
- **Group with missing flags**: If a group references a flag that doesn't exist on the current command, the entire group is skipped for that command. This is the `hasAllFlags` check.
- **Panic on missing flag**: `MarkFlagsRequiredTogether` panics if a flag name doesn't exist. Always define flags before marking groups.
- **DisableFlagParsing bypass**: If `DisableFlagParsing = true`, `ValidateFlagGroups` returns nil immediately.
- **Changed vs set**: The `pflag.Flag.Changed` field tracks whether the flag was explicitly set by the user (not just has a default value). This is what determines group validation.

## References

- [cobra source: flag_groups.go](https://github.com/spf13/cobra/blob/main/flag_groups.go)

## Related Skills

- For how `ValidateFlagGroups` fits into the execute flow, see [references/execute-flow.md](references/execute-flow.md).
- For how flag groups affect completion, see [references/completion-engine.md](references/completion-engine.md).
