---
name: cobra-dev
description: >
  Use when developing or debugging spf13/cobra internals — Command struct internals,
  ExecuteC dispatch, Find vs Traverse, flag set hierarchy, completion engine (__complete,
  getCompletions, ShellCompDirective), shell script generation (bash V1/V2, zsh, fish,
  powershell), flag group validation, global state, Levenshtein suggestions, or
  understanding the internal execution model. Triggers on: "cobra internals", "cobra ExecuteC",
  "cobra Find", "cobra Traverse", "cobra findNext", "cobra mergePersistentFlags",
  "cobra ParseFlags", "cobra getCompletions", "cobra __complete", "cobra completion engine",
  "cobra bash V2", "cobra flag resolution", "cobra flag groups internals",
  "cobra global state", "cobra EnablePrefixMatching", "cobra EnableCaseInsensitive",
  "cobra Levenshtein", "cobra stripFlags", "cobra legacyArgs", "cobra execute",
  "cobra preRun", "cobra postRun", "cobra mousetrap", "cobra templateFuncs".
user-invocable: true
---

# Cobra Development Reference

Reference for developing and debugging [cobra](https://github.com/spf13/cobra) internals. Covers the Command struct, execution flow, flag resolution, completion engine, and shell script generation.

## Data Flow

```
os.Args[1:]
  → ExecuteC()
    → Root().ExecuteC() (always starts at root)
    → InitDefaultHelpCmd, initCompleteCmd, InitDefaultCompletionCmd
    → Find(args) or Traverse(args) (based on TraverseChildren)
      → findNext() (exact match → alias → prefix match)
    → execute(flags)
      → ParseFlags
      → ValidateArgs
      → PersistentPreRun chain (parent → child)
      → PreRun
      → ValidateRequiredFlags, ValidateFlagGroups
      → Run / RunE
      → PostRun
      → PersistentPostRun chain (child → parent)
```

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

| Keywords | Reference |
|----------|----------|
| Command struct, all fields, internal state, commands slice, parent, flags, pflags, lflags, iflags, parentsPflags, FParseErrWhitelist, CompletionOptions, commandCalledAs, ctx | [references/command-struct.md](references/command-struct.md) |
| ExecuteC, execute, Find, Traverse, findNext, stripFlags, argsMinusFirstX, legacyArgs, command resolution, flag stripping, prefix matching, TraverseChildren | [references/execute-flow.md](references/execute-flow.md) |
| Flags, PersistentFlags, LocalFlags, InheritedFlags, LocalNonPersistentFlags, NonInheritedFlags, mergePersistentFlags, ParseFlags, flag set hierarchy, parentsPflags, globNormFunc, Flag lookup, persistentFlag lookup | [references/flag-resolution.md](references/flag-resolution.md) |
| __complete, __completeNoDesc, getCompletions, checkIfFlagCompletion, ShellCompDirective, CompletionFunc, RegisterFlagCompletionFunc, flagCompletionFunctions, completion output format, directive handling, ValidArgsFunction, completion flow | [references/completion-engine.md](references/completion-engine.md) |
| bash V1, bash V2, zsh, fish, powershell, GenBashCompletion, GenBashCompletionV2, GenZshCompletion, GenFishCompletion, GenPowerShellCompletion, script architecture, V1 vs V2 comparison, completion script internals | [references/shell-completions.md](references/shell-completions.md) |
| flag groups, MarkFlagsRequiredTogether, MarkFlagsMutuallyExclusive, MarkFlagsOneRequired, ValidateFlagGroups, processFlagForGroupAnnotation, annotation storage, requiredAsGroupAnnotation, oneRequiredAnnotation, mutuallyExclusiveAnnotation, enforceFlagGroupsForCompletion | [references/flag-groups-internals.md](references/flag-groups-internals.md) |
| global state, OnInitialize, OnFinalize, initializers, finalizers, EnablePrefixMatching, EnableCaseInsensitive, EnableCommandSorting, EnableTraverseRunHooks, templateFuncs, AddTemplateFunc, MousetrapHelpText, MousetrapDisplayDuration, CheckErr, Gt, Eq | [references/global-state.md](references/global-state.md) |
| suggestions, Levenshtein, findNext, prefix matching, case-insensitive, SuggestFor, SuggestionsMinimumDistance, DisableSuggestions, ld function, hasNameOrAliasPrefix, commandNameMatches | [references/suggestions.md](references/suggestions.md) |

## Quick Guide

- **How does ExecuteC dispatch to the right command?** → [references/execute-flow.md](references/execute-flow.md)
- **What is the difference between Find and Traverse?** → [references/execute-flow.md](references/execute-flow.md)
- **How does the flag set hierarchy work?** → [references/flag-resolution.md](references/flag-resolution.md)
- **How does the __complete hidden command work?** → [references/completion-engine.md](references/completion-engine.md)
- **How are shell completion scripts generated?** → [references/shell-completions.md](references/shell-completions.md)
- **How do flag group annotations work internally?** → [references/flag-groups-internals.md](references/flag-groups-internals.md)
- **What global state does cobra maintain?** → [references/global-state.md](references/global-state.md)
- **How does the suggestion/typo system work?** → [references/suggestions.md](references/suggestions.md)
- **What are all the Command struct fields?** → [references/command-struct.md](references/command-struct.md)

## Cross-Project References

- For user-facing cobra topics (creating commands, defining flags, argument validation, lifecycle hooks, help/usage, completions API, Viper integration), use the **cobra** skill.
- For pflag internals (non-POSIX modes, carapace-pflag extensions), use the **carapace-dev** skill → `references/pflag.md`.
