# Ion Shell Completion System

In-depth reference for ion's completion system — the Completer trait, IonCompleter, IonFileCompleter, MultiCompleter, and how carapace integrates with ion.

## Overview

Ion's completion system is built on the **Liner** library (`redox_liner` crate), which provides a `Completer` trait and an event-driven completion pipeline. Unlike bash, fish, or zsh, ion has **no user-facing API for custom completions** — all completion logic is compiled into the shell. There is no equivalent to bash's `complete` builtin, fish's `complete` command, or zsh's `compadd`.

The completion pipeline:

1. User presses **Tab** (Ctrl+i)
2. Liner fires `BeforeComplete` event
3. `IonCompleter.on_event()` determines `CompletionType` from cursor position
4. `IonCompleter.completions()` dispatches to the appropriate handler
5. Results are sorted, deduplicated, and displayed

## CompletionType Enum

Ion classifies completion context into three types:

```rust
enum CompletionType {
    Nothing,           // No completion available
    Command,           // Completing command names
    VariableAndFiles,  // Completing variables and file paths
}
```

### Context Detection (on_event)

The `CompletionType` is determined in `IonCompleter.on_event()` based on cursor position:

| CursorPosition | Word Index | CompletionType |
|----------------|------------|----------------|
| `InWord(0)` | First word | `Command` |
| `OnWordRightEdge(0)` | First word | `Command` |
| `OnWordRightEdge(n)` where word n-1 ends with `\|`, `&`, or `;` | After pipe/bg/seq | `Command` |
| All other positions | — | `VariableAndFiles` |
| Empty words | — | `Nothing` |

The pipe/background detection checks if the **previous word** ends with `|`, `&`, or `;`:

```rust
let is_pipe = words
    .into_iter()
    .nth(index - 1)
    .map(|(start, end)| event.editor.current_buffer().range(start, end))
    .filter(|filename| {
        filename.ends_with('|')
            || filename.ends_with('&')
            || filename.ends_with(';')
    })
    .is_some();
```

## The Completer Trait

From the `redox_liner` crate:

```rust
pub trait Completer {
    fn completions(&mut self, start: &str) -> Vec<String>;
    fn on_event<W: Write>(&mut self, _event: Event<'_, '_, W>) { ... }
}
```

- `completions(start)` — given a partial word, return matching candidates
- `on_event(event)` — optional hook fired before completion processing

The trait is **not dyn-compatible** (not object-safe). It is used as a generic parameter to `read_line()`.

## IonCompleter

The main completion struct:

```rust
pub struct IonCompleter<'a, 'b> {
    shell:      &'b Shell<'a>,
    completion: CompletionType,
}
```

### completions() Dispatch

```rust
fn completions(&mut self, start: &str) -> Vec<String> {
    let mut completions = IonFileCompleter::new(None, &self.shell).completions(start);
    let vars = self.shell.variables();

    match self.completion {
        CompletionType::VariableAndFiles => { /* variables + files */ }
        CompletionType::Command => { /* builtins + aliases + functions + PATH */ }
        CompletionType::Nothing => (),
    }
    completions
}
```

**Key detail**: File completions are **always** included as the base set, even for `VariableAndFiles` and `Command` types. The `IonFileCompleter::new(None, &self.shell)` call (with `None` path) uses the current directory.

### Command Completion

When `CompletionType::Command`:

1. **Builtins** — `self.shell.builtins().keys()`
2. **Aliases** — `vars.aliases().map(|(key, _)| key.to_string())`
3. **Functions** — `vars.functions().map(|(key, _)| key.to_string())`
4. **PATH executables** — scans each directory in `$PATH` using `IonFileCompleter` with `for_command: true`

PATH scanning creates one `IonFileCompleter` per PATH directory and merges results with `MultiCompleter`:

```rust
let file_completers: Vec<_> = if let Some(paths) = env::var_os("PATH") {
    env::split_paths(&paths)
        .map(|s| {
            let s = if !s.to_string_lossy().ends_with('/') {
                let mut oss = s.into_os_string();
                oss.push("/");
                oss.into()
            } else { s };
            IonFileCompleter::new(Some(s), &self.shell)
        })
        .collect()
} else {
    vec![IonFileCompleter::new(Some("/bin/".into()), &self.shell)]
};
completions.extend(MultiCompleter::new(file_completers).completions(start));
```

When `for_command` is true, `IonFileCompleter` strips the directory prefix from results, showing only the filename.

### Variable Completion

When `CompletionType::VariableAndFiles`:

| Input prefix | Completions |
|-------------|-------------|
| Empty string | All `$string_vars` + all `@arrays` |
| Starts with `$` | String variables matching after `$` |
| Starts with `@` | Array variables matching after `@` |
| Other | File completions only (from base set) |

```rust
if start.is_empty() {
    completions.extend(vars.string_vars().map(|(s, _)| format!("${}", s)));
    completions.extend(vars.arrays().map(|(s, _)| format!("@{}", s)));
} else if start.starts_with('$') {
    completions.extend(
        vars.string_vars()
            .filter(|(s, _)| s.starts_with(&start[1..]))
            .map(|(s, _)| format!("${}", &s)),
    );
} else if start.starts_with('@') {
    completions.extend(
        vars.arrays()
            .filter(|(s, _)| s.starts_with(&start[1..]))
            .map(|(s, _)| format!("@{}", &s)),
    );
}
```

## IonFileCompleter

Wraps file completion with ion-specific handling:

```rust
pub struct IonFileCompleter<'a, 'b> {
    shell:       &'b Shell<'a>,
    path:        PathBuf,    // Working directory for glob
    for_command: bool,       // True: strip path prefix, show only filename
}
```

### Tilde Expansion

`IonFileCompleter` calls `self.shell.tilde(start)` to expand `~` to the home directory before glob matching. The tilde prefix is preserved in the output:

```rust
let expanded = match self.shell.tilde(start) {
    Ok(expanded) => expanded,
    Err(why) => { eprintln!("ion: {}", why); return vec![start.into()]; }
};
```

### filename_completion()

Uses the `glob` crate with `MatchOptions`:

```rust
fn filename_completion<'a>(start: &'a str, path: &Path) -> impl Iterator<Item = String> + 'a {
    let unescaped_start = unescape(start);
    // Build glob pattern: path/partial*
    // Uses glob_with with:
    //   case_sensitive: true
    //   require_literal_separator: true
    //   require_literal_leading_dot: false
}
```

- Appends `*` to the partial word for glob matching
- Adds trailing `/` for directories
- Handles absolute paths (starts with `/`)
- Handles `./` prefix

### Escape/Unescape

Ion escapes special shell characters in completion results:

**Characters that get backslash-escaped:**

```
( ) [ ] & $ @ { } < > ; " ' # ^ * <space>
```

```rust
fn escape(input: &str) -> String {
    for character in input.bytes() {
        match character {
            b'(' | b')' | b'[' | b']' | b'&' | b'$' | b'@' | b'{' | b'}'
            | b'<' | b'>' | b';' | b'"' | b'\'' | b'#' | b'^' | b'*' | b' '
            => output.push(b'\\'),
            _ => (),
        }
        output.push(character);
    }
}
```

`unescape()` reverses this for the input before glob matching.

## MultiCompleter

Combines results from multiple completers:

```rust
pub struct MultiCompleter<A>(Vec<A>);

impl<A: Completer> Completer for MultiCompleter<A> {
    fn completions(&mut self, start: &str) -> Vec<String> {
        self.0.iter_mut()
            .flat_map(|comp| comp.completions(start))
            .collect()
    }
}
```

Used for PATH scanning — one `IonFileCompleter` per directory, merged with `MultiCompleter`.

## Completion Display

Liner handles completion display in the `Editor.complete()` method:

1. **0 matches** — do nothing
2. **1 match** — auto-insert (delete current word, insert completion)
3. **Multiple matches with common prefix** — auto-insert the common prefix
4. **Multiple matches, no common prefix** — store in `show_completions_hint` and display list

Subsequent Tab presses **cycle** through the completion list (not menu-complete like bash).

The completion list is displayed in columns by `print_completion_list()`:

```rust
fn print_completion_list(completions: &[String], highlighted: Option<usize>, output_buf: &mut String) {
    let (w, _) = termion::terminal_size()?;
    let max_word_size = completions.iter().fold(1, |m, x| max(m, x.chars().count()));
    let cols = max(1, w as usize / (max_word_size));
    let col_width = 2 + w as usize / cols;
    // Print in columns, highlight selected with black-on-white
}
```

- Multi-column layout based on terminal width
- Currently selected completion highlighted with **black text on white background**
- No description support — completions are plain strings

## Nospace Handling

Ion has **no native nospace mechanism**. The Liner library always inserts the completion as-is. Directories get a trailing `/` appended by `filename_completion()`, but there is no per-candidate suffix control like bash BLE's suffix field or elvish's `CodeSuffix`.

## Carapace Integration

### Source Files

| File | Purpose |
|------|---------|
| `internal/shell/ion/action.go` | `ActionRawValues()` — formats completion output as JSON |
| `internal/shell/ion/snippet.go` | `Snippet()` — returns empty string (no snippet) |

### The Snippet

Ion's `Snippet()` returns an **empty string**. Ion has no mechanism to register external completion functions — there is no `complete` builtin, no `compdef`, no `complete -c`. Carapace cannot generate a setup script for ion.

This means carapace's ion support is **experimental** — it can format completions for ion, but there is no standard way to hook them into the shell.

### Value Formatting (ActionRawValues)

Carapace formats ion completions as a JSON array of suggestion objects:

```go
type suggestion struct {
    Value   string
    Display string
}
```

**Format rules:**

- Newlines and carriage returns are stripped from Value, Display, and Description
- If `meta.Nospace` matches the value, **no trailing space** is appended
- Otherwise, a trailing space is appended to `Value`
- If a description exists, Display is formatted as `value (description)`
- If no description, Display equals the raw display value

**Example output:**

```json
[
  {"Value": "add ", "Display": "add"},
  {"Value": "commit ", "Display": "commit (record changes)"},
  {"Value": "push", "Display": "push"}
]
```

The trailing space in `Value` implements nospace handling — when `Nospace` matches, the space is omitted.

### Comparison with Other Shells

| Feature | Ion | Bash | Fish | Zsh |
|---------|-----|------|------|-----|
| Custom completion API | None | `complete` builtin | `complete` command | `compadd`/`_describe` |
| Registration mechanism | None | compspec | `complete -c` | `compdef` |
| Per-candidate nospace | Trailing space in Value | `compopt -o nospace` (global) | Not supported | Space suffix in values |
| Description display | Inline `(desc)` | COMP_TYPE=63 only | Pager column | Right column |
| Style/color | Not supported | Not supported | Not supported | `zstyle list-colors` |
| Snippet | Empty | Full setup script | Full setup script | Full setup script |
| Go-side patching | None | `bash.Patch()` | None | None |

## Limitations and Known Issues

1. **No custom completion API** — ion cannot register per-command completions. All completion logic is hardcoded in `completer.rs`.
2. **No description support** — the Liner `Completer` trait returns `Vec<String>`, not structured candidates with descriptions.
3. **No style/color support** — completions are plain strings with no styling.
4. **No tag grouping** — all completions are merged into a single list.
5. **No message support** — there is no way to display informational messages during completion.
6. **No argument-aware completion** — ion cannot complete command arguments differently based on which command is being invoked. All non-first-word completions are file/variable completions.
7. **No plugin completion** — ion-plugins can only define aliases, functions, and environment variables, not custom completion logic.
8. **Cycling behavior** — subsequent Tab presses cycle through completions one at a time, rather than showing a menu.

## Extending Ion Completions

The only way to add custom completions to ion is to modify the Rust source code:

1. Add a new `CompletionType` variant (e.g., `CommandArgs`)
2. Detect the command context in `on_event()` (check the first word)
3. Implement the completion logic in `completions()`
4. Recompile ion

Alternatively, a future ion version could expose a completion API similar to other shells, but this does not exist as of the current release.

## References

- [Ion completer.rs source](https://github.com/redox-os/ion/blob/master/src/binary/completer.rs)
- [redox_liner Completer trait](https://docs.rs/redox_liner/latest/linter/trait.Completer.html)
- [redox_liner Event/EventKind](https://docs.rs/redox_liner/latest/linter/struct.Event.html)
- [redox_liner CursorPosition](https://docs.rs/redox_liner/latest/linter/enum.CursorPosition.html)
- [carapace ion action.go](https://github.com/carapace-sh/carapace/blob/master/internal/shell/ion/action.go)

## Related Skills

- [references/line-editing.md](references/line-editing.md) — the Liner library that powers ion's line editing and completion display
- [references/language.md](references/language.md) — ion syntax including quoting rules that affect completion
- **carapace-dev** skill → `references/shell.md` — cross-shell comparison including ion
