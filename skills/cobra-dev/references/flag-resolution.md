# Flag Set Hierarchy and Resolution

How cobra's flag sets relate to each other, how they're merged, and how `ParseFlags` works.

> **Source of truth**: <https://github.com/spf13/cobra/blob/main/command.go>. For user-facing flag usage, see the **cobra** skill → `references/flags.md`.

## The Five Flag Sets

Each `Command` has five flag sets, each serving a different purpose:

| Field | Method | Scope | Lazy |
|-------|--------|-------|------|
| `pflags` | `PersistentFlags()` | Flags declared as persistent on this command | Yes |
| `flags` | `Flags()` | Merged set: local + persistent + inherited | Yes |
| `lflags` | `LocalFlags()` | Flags local to this command (not inherited) | Yes |
| `iflags` | `InheritedFlags()` | Flags inherited from parents only | Yes |
| `parentsPflags` | (no public method) | All persistent flags from the parent chain | No (set by `mergePersistentFlags`) |

### Hierarchy Diagram

```
Root.PersistentFlags()     →  --verbose, --config
Root.Flags()               →  --verbose, --config, --help (root local)

  Sub.PersistentFlags()    →  --output
  Sub.Flags()              →  --verbose, --config, --output, --format (sub local)
  Sub.LocalFlags()         →  --output, --format (not inherited)
  Sub.InheritedFlags()     →  --verbose, --config (from root)
  Sub.parentsPflags        →  --verbose, --config (root's persistent flags)
```

## Flags() — The Merged Set

```go
func (c *Command) Flags() *flag.FlagSet {
    if c.flags == nil {
        c.flags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
        c.flags.SetOutput(c.flagErrorBuf)
    }
    return c.flags
}
```

`Flags()` returns the **merged** flag set. The merge happens in `mergePersistentFlags()` which is called by `LocalFlags()` and `InheritedFlags()`.

## PersistentFlags() — This Command's Persistent Flags

```go
func (c *Command) PersistentFlags() *flag.FlagSet {
    if c.pflags == nil {
        c.pflags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
        c.pflags.SetOutput(c.flagErrorBuf)
    }
    return c.pflags
}
```

These flags cascade to all children. They are stored in `c.pflags` and also added to `c.flags` during merge.

## LocalFlags() — Non-Inherited Flags

```go
func (c *Command) LocalFlags() *flag.FlagSet {
    c.mergePersistentFlags()  // ensure parentsPflags is populated

    if c.lflags == nil {
        c.lflags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
        c.lflags.SetOutput(c.flagErrorBuf)
    }

    addToLocal := func(f *flag.Flag) {
        // Add if not already in lflags AND not in parentsPflags
        if c.lflags.Lookup(f.Name) == nil && f != c.parentsPflags.Lookup(f.Name) {
            c.lflags.AddFlag(f)
        }
    }
    c.Flags().VisitAll(addToLocal)
    c.PersistentFlags().VisitAll(addToLocal)
    return c.lflags
}
```

**Key logic**: A flag is "local" if it's in `Flags()` or `PersistentFlags()` but **not** in `parentsPflags`. This means a persistent flag declared on this command is local here, but inherited in children.

## InheritedFlags() — From Parents Only

```go
func (c *Command) InheritedFlags() *flag.FlagSet {
    c.mergePersistentFlags()

    if c.iflags == nil {
        c.iflags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
        c.iflags.SetOutput(c.flagErrorBuf)
    }

    local := c.LocalFlags()
    c.parentsPflags.VisitAll(func(f *flag.Flag) {
        // Add if not already in iflags AND not in local
        if c.iflags.Lookup(f.Name) == nil && local.Lookup(f.Name) == nil {
            c.iflags.AddFlag(f)
        }
    })
    return c.iflags
}
```

**Key logic**: Inherited flags are those in `parentsPflags` that are **not** shadowed by a local flag. If a child declares a local flag with the same name, the local flag wins.

## mergePersistentFlags — Building parentsPflags

```go
func (c *Command) mergePersistentFlags() {
    c.parentsPflags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
    c.parentsPflags.SetOutput(c.flagErrorBuf)

    // Walk up the parent chain
    for p := c; p != nil; p = p.Parent() {
        if p.PersistentFlags() == nil {
            continue
        }
        p.PersistentFlags().VisitAll(func(f *flag.Flag) {
            // Don't overwrite: first (closest) parent wins
            if c.parentsPflags.Lookup(f.Name) == nil {
                c.parentsPflags.AddFlag(f)
            }
        })
    }
    // Also merge parentsPflags into c.flags
    c.parentsPflags.VisitAll(func(f *flag.Flag) {
        if c.Flags().Lookup(f.Name) == nil {
            c.Flags().AddFlag(f)
        }
    })
}
```

**Key detail**: When walking up the parent chain, the **closest parent wins**. If both root and a middle command define a persistent flag with the same name, the middle command's version is used (because it's visited first).

**Side effect**: `mergePersistentFlags` also populates `c.flags` with inherited persistent flags. This is why `Flags()` returns the full merged set.

## ParseFlags — How Flags Are Parsed

```go
func (c *Command) ParseFlags(args []string) error {
    // First, merge inherited persistent flags into c.flags
    c.mergePersistentFlags()

    // Apply global normalization function
    if c.globNormFunc != nil {
        c.Flags().SetNormalizeFunc(c.globNormFunc)
    }

    // Enable error handling
    c.Flags().SetOutput(c.flagErrorBuf)

    // Parse
    err := c.Flags().Parse(args)

    // Restore output
    c.Flags().SetOutput(nil)

    return err
}
```

### Parse Order

1. `mergePersistentFlags()` — populate `c.flags` with inherited persistent flags
2. Apply normalization function
3. `c.Flags().Parse(args)` — pflag parses the merged set

This means **all** flags (local, persistent, inherited) are parsed in a single pass.

## Flag Lookup Methods

### Flag(name) — Climb the Tree

```go
func (c *Command) Flag(name string) *flag.Flag {
    // Search from current command up to root
    // Checks c.Flags() (merged set) at each level
}
```

### persistentFlag(name) — Recursive Persistent Search

```go
func (c *Command) persistentFlag(name string) *flag.Flag {
    // Search c.PersistentFlags() first
    // Then recurse to parent.PersistentFlags()
}
```

## Flag Shadowing Rules

When a child defines a flag with the same name as a parent's persistent flag:

1. **LocalFlags()**: The child's version is included (it's not in `parentsPflags`)
2. **InheritedFlags()**: The parent's version is **excluded** (shadowed by local)
3. **Flags()**: The child's version is in the merged set; the parent's is not (first-write wins during merge)
4. **ParseFlags**: The child's version is parsed (it's in the merged set)

## Boolean Check Methods

| Method | Checks |
|--------|--------|
| `HasFlags()` | `c.flags != nil && c.flags.HasFlags()` |
| `HasPersistentFlags()` | `c.pflags != nil && c.pflags.HasFlags()` |
| `HasLocalFlags()` | `c.lflags != nil && c.lflags.HasFlags()` |
| `HasInheritedFlags()` | `c.iflags != nil && c.iflags.HasFlags()` |
| `HasAvailableFlags()` | `HasFlags() || HasInheritedFlags()` |
| `HasAvailablePersistentFlags()` | Persistent flags that are not hidden |
| `HasAvailableLocalFlags()` | Local flags that are not hidden |
| `HasAvailableInheritedFlags()` | Inherited flags that are not hidden |

## Edge Cases

- **Lazy initialization**: All flag sets are lazily created. Calling `Flags()` on a command that has never had flags defined still returns a valid (empty) `FlagSet`.
- **mergePersistentFlags is called often**: Both `LocalFlags()` and `InheritedFlags()` call it, which rebuilds `parentsPflags` each time. This is not cached between calls.
- **globNormFunc propagation**: When `AddCommand` is called, the parent's `globNormFunc` is propagated to the child via `SetGlobalNormalizationFunc`. This ensures consistent flag name normalization across the tree.
- **flagErrorBuf sharing**: `flagErrorBuf` is shared between `flags`, `pflags`, `lflags`, and `iflags` — they all write parse errors to the same buffer.

## References

- [cobra source: command.go](https://github.com/spf13/cobra/blob/main/command.go) — flag methods (lines 1688–1900)
- [pflag source](https://github.com/spf13/pflag) — underlying flag set implementation

## Related Skills

- For how `ParseFlags` fits into the execute flow, see [references/execute-flow.md](references/execute-flow.md).
- For pflag non-POSIX extensions, see the **carapace-dev** skill → `references/pflag.md`.
