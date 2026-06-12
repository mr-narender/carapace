# Shell Completion Script Generation

How cobra generates completion scripts for bash, zsh, fish, and PowerShell.

> **Source of truth**: <https://github.com/spf13/cobra/blob/main/bash_completions.go>, <https://github.com/spf13/cobra/blob/main/bash_completionsV2.go>, <https://github.com/spf13/cobra/blob/main/zsh_completions.go>, <https://github.com/spf13/cobra/blob/main/fish_completions.go>, <https://github.com/spf13/cobra/blob/main/powershell_completions.go>. For user-facing completion setup, see the **cobra** skill → `references/completions.md`.

## Architecture Overview

All shells follow the same pattern:

1. **Generate a shell script** that registers a completion function
2. The completion function calls the program's `__complete` hidden command
3. The program returns completions + a directive integer
4. The script interprets the directive and presents completions

The key difference between shells is how they handle the output.

## Bash V1 (bash_completions.go)

### Public API

```go
func (c *Command) GenBashCompletion(w io.Writer) error
func (c *Command) GenBashCompletionFile(filename string) error
```

### Architecture

V1 generates a **complete, standalone bash script** that contains all command structure at generation time:

```
# Preamble: helper functions
__<prog>_debug()
__<prog>_init_completion()
__<prog>_handle_go_custom_completion()
__<prog>_handle_reply()
__<prog>_handle_flag()
__<prog>_handle_noun()
__<prog>_handle_command()
__<prog>_handle_word()

# Per-command functions (recursively generated)
_<prog>()
_<prog>_sub1()
_<prog>_sub1_sub2()

# Postscript: entry point
__start_<prog>()
complete -o default -F __start_<prog> <prog>
```

### Per-Command Function

Each `_<command_path>()` function contains:

| Section | Content |
|---------|---------|
| `writeCommands()` | Subcommand names |
| `writeFlags()` | Flag names, two-word flags, local non-persistent flags |
| `writeRequiredFlag()` | Required flags (BashCompOneRequiredFlag) |
| `writeRequiredNouns()` | ValidArgs values (descriptions stripped) |
| `writeArgAliases()` | Argument aliases |
| `writeFlagHandler()` | Custom flag completion handlers (BashCompCustom) |

### Limitations

- **No descriptions**: ValidArgs descriptions (tab-separated) are stripped
- **No ActiveHelp**: Sets `COBRA_ACTIVE_HELP=0` to disable it
- **No KeepOrder**: Only handles 5 directives (Error, NoSpace, NoFileComp, FilterFileExt, FilterDirs)
- **Large output**: One function per command/subcommand
- **Must regenerate**: Changes to CLI structure require regenerating the script

## Bash V2 (bash_completionsV2.go)

### Public API

```go
func (c *Command) GenBashCompletionV2(w io.Writer, includeDesc bool) error
func (c *Command) GenBashCompletionFileV2(filename string, includeDesc bool) error
```

### Architecture

V2 generates a **thin wrapper** that delegates all logic to the Go binary:

```
__<prog>_debug()
__<prog>_init_completion()
__<prog>_get_completion_results()    # calls <prog> __complete
__<prog>_process_completion_results() # interprets directives
__<prog>_extract_activeHelp()
__<prog>_handle_completion_types()
__<prog>_format_comp_descriptions()
__<prog>_handle_special_char()
__<prog>_reprint_commandLine()
__start_<prog>()
complete -o default -F __start_<prog> <prog>
```

### Key Functions

**`__<prog>_get_completion_results`**:
- Builds the command line: `<prog> __complete <args>`
- Adds an empty parameter if cursor is at end (to signal "completing empty word")
- Handles `--flag=` prefix completion
- Calls `eval` to execute the program
- Splits output on last `:` to extract directive

**`__<prog>_process_completion_results`**:
- Checks each directive bit (Error, NoSpace, NoFileComp, FilterFileExt, FilterDirs, KeepOrder)
- Applies `compopt` settings (nospace, nosort, +o default)
- Handles file extension filtering via `_filedir`
- Handles directory filtering via `_filedir -d`
- Extracts and displays ActiveHelp messages

**`__<prog>_format_comp_descriptions`**:
- Aligns completion descriptions in columns
- Trims descriptions to fit terminal width

### V1 vs V2 Comparison

| Aspect | V1 | V2 |
|--------|----|----|
| Architecture | Generated shell script with hardcoded structure | Thin wrapper delegating to Go binary |
| Output size | Large (one function per command) | Small (single function) |
| Descriptions | Not supported | Fully supported |
| ActiveHelp | Not supported (disabled) | Fully supported |
| KeepOrder | Not supported | Supported (bash ≥4.4) |
| Flag parsing | Done in shell script | Done by Go binary |
| Performance | Faster (no fork/exec per TAB) | Slower (fork/exec per TAB) |
| Maintenance | Must regenerate on changes | Always up-to-date |

## Zsh (zsh_completions.go)

### Public API

```go
func (c *Command) GenZshCompletion(w io.Writer) error
func (c *Command) GenZshCompletionNoDesc(w io.Writer) error
func (c *Command) GenZshCompletionFile(filename string) error
func (c *Command) GenZshCompletionFileNoDesc(filename string) error
```

### Architecture

Single function `_<prog>()` that:

1. Calls `<prog> __complete <args>` via `eval`
2. Parses directive from last line of output
3. Handles directives:
   - `FilterFileExt` → `_files -g` with glob patterns
   - `FilterDirs` → `_arguments '*:dirname:_files -/'`
   - Default → `_describe` for completions with descriptions
   - `NoFileComp` not set → fall back to `_arguments '*:filename:_files'`
4. `KeepOrder` → uses `-V` flag for `_describe`
5. `NoSpace` → uses `-S ''` flag
6. ActiveHelp → `compadd -x` for help text display

### Description Control

- `includeDesc = true` → uses `__complete`
- `includeDesc = false` → uses `__completeNoDesc`, strips tab+description

## Fish (fish_completions.go)

### Public API

```go
func (c *Command) GenFishCompletion(w io.Writer, includeDesc bool) error
func (c *Command) GenFishCompletionFile(filename string, includeDesc bool) error
```

### Architecture

Multiple helper functions:

| Function | Purpose |
|----------|---------|
| `__<prog>_debug` | Debug logging |
| `__<prog>_perform_completion` | Call `__complete`, capture output, extract directive |
| `__<prog>_perform_completion_once` | Cache completion results |
| `__<prog>_clear_perform_completion_once_result` | Clear cache |
| `__<prog>_requires_order_preservation` | Check KeepOrder directive |
| `__<prog>_prepare_completions` | Main completion logic |

### Fish-Specific Handling

- **NoSpace trick**: Adds an extra longer completion to prevent fish from adding a space
- **File completion fallback**: Counts completions; if none, falls back to file completion
- **Order preservation**: Uses `-k` flag on `complete` when KeepOrder is set
- **Flag prefix**: Handles `--flag=` syntax for flag value completion

## PowerShell (powershell_completions.go)

### Public API

```go
func (c *Command) GenPowerShellCompletion(w io.Writer) error
func (c *Command) GenPowerShellCompletionWithDesc(w io.Writer) error
func (c *Command) GenPowerShellCompletionFile(filename string) error
func (c *Command) GenPowerShellCompletionFileWithDesc(filename string) error
```

### Architecture

Creates a `[scriptblock]` registered with `Register-ArgumentCompleter`:

1. Extracts command line from `$CommandAst` and `$CursorPosition`
2. Truncates to cursor position
3. Calls `<prog> __complete <args>` via `Invoke-Expression`
4. Parses directive from last line
5. Splits completions on tab for Name/Description
6. Handles directives:
   - `Error` → return empty
   - `NoSpace` → set `CompletionResult` without trailing space
   - `NoFileComp` → return without file completion
   - `FilterFileExt`/`FilterDirs` → not natively supported, return empty
   - `KeepOrder` → skip `Sort-Object`
7. Supports three completion modes: `Complete`, `MenuComplete`, `TabCompleteNext`
8. Escapes special characters via `__<prog>_escapeStringWithSpecialChars`

### Requirements

PowerShell v5.0+ required.

## Common Patterns Across Shells

All shell scripts share these patterns:

1. **Call the program** with `__complete` (or `__completeNoDesc`)
2. **Parse the directive** from the last line (`:<integer>`)
3. **Handle directives** with shell-specific mechanisms
4. **ActiveHelp** entries (prefixed with `_activeHelp_ `) are displayed as messages, not completions
5. **File completion fallback** when no completions are returned and `NoFileComp` is not set

## Edge Cases

- **Bash V1 + custom completions**: V1 uses `BashCompCustom` annotations to call `__<prog>_handle_go_custom_completion`, which invokes the Go binary. This is a hybrid approach within V1.
- **Fish NoSpace**: Fish doesn't have a native nospace mechanism. The workaround adds an extra longer completion that prevents the space from being inserted.
- **PowerShell FilterFileExt/FilterDirs**: Not natively supported in PowerShell's completion system. These directives are effectively no-ops.
- **Zsh _describe fallback**: If `_describe` returns no matches, zsh falls back to file completion (unless `NoFileComp` is set).

## References

- [cobra source: bash_completions.go](https://github.com/spf13/cobra/blob/main/bash_completions.go)
- [cobra source: bash_completionsV2.go](https://github.com/spf13/cobra/blob/main/bash_completionsV2.go)
- [cobra source: zsh_completions.go](https://github.com/spf13/cobra/blob/main/zsh_completions.go)
- [cobra source: fish_completions.go](https://github.com/spf13/cobra/blob/main/fish_completions.go)
- [cobra source: powershell_completions.go](https://github.com/spf13/cobra/blob/main/powershell_completions.go)

## Related Skills

- For the `__complete` command and `getCompletions` internals, see [references/completion-engine.md](references/completion-engine.md).
- For in-depth bash completion knowledge, use the **bash** skill.
- For in-depth zsh completion knowledge, use the **zsh** skill.
- For in-depth fish completion knowledge, use the **fish** skill.
- For in-depth PowerShell completion knowledge, use the **powershell** skill.
