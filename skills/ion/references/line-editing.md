# Ion Shell Line Editing (Liner Library)

In-depth reference for ion's line editing system — the redox_liner crate, Editor, Context, KeyMap, completion display, autosuggestions, and history integration.

## Overview

Ion uses the **redox_liner** crate (a Rust readline replacement) for all line editing functionality. This library provides:

- Emacs and Vi editing modes
- Tab completion with the `Completer` trait
- Fish-like autosuggestions from history
- Incremental history search (Ctrl+R)
- Multi-column completion display
- Customizable word boundary detection

## The Context Struct

`Context` is the main configuration object for the line editor:

```rust
pub struct Context {
    pub history: History,
    pub word_divider_fn: Box<dyn Fn(&Buffer) -> Vec<(usize, usize)>>,
    pub key_bindings: KeyBindings,
    pub buf: String,
}

pub enum KeyBindings {
    Vi,
    Emacs,
}
```

| Field | Purpose |
|-------|---------|
| `history` | Command history storage and retrieval |
| `word_divider_fn` | Function that splits the buffer into word boundaries |
| `key_bindings` | Active keybinding mode (Vi or Emacs) |
| `buf` | Internal buffer |

### Creating a Context

```rust
let mut context = Context::new();
context.key_bindings = KeyBindings::Vi;  // or KeyBindings::Emacs (default)
```

### Word Division

The default word divider (`get_buffer_words`) splits on unescaped spaces:

```rust
pub fn get_buffer_words(buf: &Buffer) -> Vec<(usize, usize)> {
    let mut res = Vec::new();
    let mut word_start = None;
    let mut just_had_backslash = false;

    for (i, &c) in buf.chars().enumerate() {
        if c == '\\' {
            just_had_backslash = true;
            continue;
        }
        if let Some(start) = word_start {
            if c == ' ' && !just_had_backslash {
                res.push((start, i));
                word_start = None;
            }
        } else {
            if c != ' ' {
                word_start = Some(i);
            }
        }
        just_had_backslash = false;
    }
    if let Some(start) = word_start {
        res.push((start, buf.num_chars()));
    }
    res
}
```

- Returns `Vec<(usize, usize)>` — pairs of (start_index, end_index) for each word
- Handles backslash-escaped spaces: `hello\ world` is one word
- Customizable via `Context.word_divider_fn`

## The Editor Struct

`Editor` is the core line editor that handles all editing operations:

```rust
pub struct Editor<'a, W: io::Write> {
    prompt: String,
    out: W,
    context: &'a mut Context,
    closure: Option<ColorClosure>,
    cursor: usize,
    new_buf: Buffer,
    hist_buf: Buffer,
    hist_buf_valid: bool,
    cur_history_loc: Option<usize>,
    term_cursor_line: usize,
    show_completions_hint: Option<(Vec<String>, Option<usize>)>,
    show_autosuggestions: bool,
    pub no_eol: bool,
}
```

| Field | Purpose |
|-------|---------|
| `prompt` | The prompt string |
| `out` | Output writer (stdout) |
| `context` | Reference to Context (history, keybindings) |
| `cursor` | Current cursor position |
| `new_buf` | Current buffer being edited |
| `hist_buf` | Buffer for history navigation |
| `show_completions_hint` | Active completion list and selected index |
| `show_autosuggestions` | Whether autosuggestions are enabled |

## The KeyMap Trait

```rust
pub trait KeyMap: Default {
    fn handle_key_core<'a, W: Write>(
        &mut self,
        key: Key,
        editor: &mut Editor<'a, W>
    ) -> Result<()>;

    fn init<'a, W: Write>(&mut self, _editor: &mut Editor<'a, W>) { }

    fn handle_key<'a, W: Write, C: Completer>(
        &mut self,
        key: Key,
        editor: &mut Editor<'a, W>,
        handler: &mut C
    ) -> Result<bool> { ... }
}
```

The default `handle_key` implementation handles two keys universally:

```rust
match key {
    Key::Ctrl('i') => { editor.complete(handler)?; Ok(false) }  // Tab
    Key::Char('\n') | Key::Ctrl('m') => {                       // Enter
        if editor.handle_newline()? { Ok(true) } else { Ok(false) }
    }
    _ => self.handle_key_core(key, editor).map(|_| false),
}
```

All other keys are dispatched to `handle_key_core`, which is implemented separately for Emacs and Vi modes.

## read_line() Flow

```rust
pub fn read_line_with_init_buffer<P, B, C>(
    &mut self,
    prompt: P,
    handler: &mut C,
    f: Option<ColorClosure>,
    buffer: B,
) -> io::Result<String> {
    let stdout = stdout().into_raw_mode()?;
    let keybindings = self.key_bindings;
    let ed = Editor::new_with_init_buffer(stdout, prompt, f, self, buffer)?;
    match keybindings {
        KeyBindings::Emacs => Self::handle_keys(keymap::Emacs::new(), ed, handler),
        KeyBindings::Vi => Self::handle_keys(keymap::Vi::new(), ed, handler),
    }
}
```

### Key Handling Loop

```rust
fn handle_keys<'a, W: Write, M: KeyMap, C: Completer>(
    mut keymap: M,
    mut ed: Editor<'a, W>,
    handler: &mut C,
) -> io::Result<String> {
    keymap.init(&mut ed);
    for c in stdin().keys() {
        if keymap.handle_key(c.unwrap(), &mut ed, handler)? {
            break;  // Return true means "done editing"
        }
    }
    Ok(ed.into())
}
```

## Completion Processing

### The complete() Method

```rust
pub fn complete<T: Completer>(&mut self, handler: &mut T) -> io::Result<()> {
    // 1. Fire BeforeComplete event
    handler.on_event(Event::new(self, EventKind::BeforeComplete));

    // 2. If already cycling, advance to next completion
    if let Some((completions, i_in)) = self.show_completions_hint.take() {
        let i = i_in.map_or(0, |i| (i + 1) % completions.len());
        // Delete current word and insert next completion
        ...
        self.show_completions_hint = Some((completions, Some(i)));
        return Ok(());
    }

    // 3. Get word before cursor
    let word = self.get_word_before_cursor(true);

    // 4. Get completions
    let mut completions = handler.completions(word.as_ref());

    // 5. Sort and deduplicate
    completions.sort();
    completions.dedup();

    // 6. Handle results
    match completions.len() {
        0 => { /* do nothing */ }
        1 => { /* auto-insert single match */ }
        _ => {
            // Find longest common prefix
            let prefix = util::find_longest_common_prefix(&completions);
            if prefix.len() > word.len() {
                // Auto-insert common prefix
            } else {
                // Store completions for cycling
                self.show_completions_hint = Some((completions, None));
            }
        }
    }
}
```

### Completion Cycling

When multiple completions are available:

1. **First Tab** — displays the completion list and stores it in `show_completions_hint`
2. **Subsequent Tabs** — cycles through completions one at a time
3. The currently selected completion is highlighted with **black text on white background**

This is different from bash's menu-complete (which inserts the completion) and zsh's menu-select (which provides a navigable menu).

### CursorPosition

The `CursorPosition` enum determines what part of the line the cursor is in:

```rust
pub enum CursorPosition {
    InWord(usize),                    // Cursor is inside word at index
    OnWordLeftEdge(usize),            // Cursor is at the start of word at index
    OnWordRightEdge(usize),           // Cursor is at the end of word at index
    InSpace(Option<usize>, Option<usize>),  // Cursor is between words
}
```

**How it's computed:**

```rust
pub fn get(cursor: usize, words: &[(usize, usize)]) -> CursorPosition {
    for (i, &(start, end)) in words.iter().enumerate() {
        if start == cursor { return OnWordLeftEdge(i); }
        else if end == cursor { return OnWordRightEdge(i); }
        else if start < cursor && cursor < end { return InWord(i); }
        else if cursor < start { return InSpace(Some(i - 1), Some(i)); }
    }
    InSpace(Some(words.len() - 1), None)
}
```

### get_word_before_cursor()

```rust
fn get_word_before_cursor(&self, ignore_space_before_cursor: bool) -> Option<(usize, usize)> {
    let (words, pos) = self.get_words_and_cursor_position();
    match pos {
        CursorPosition::InWord(i) => Some(words[i]),
        CursorPosition::InSpace(Some(i), _) => {
            if ignore_space_before_cursor { Some(words[i]) } else { None }
        },
        CursorPosition::InSpace(None, _) => None,
        CursorPosition::OnWordLeftEdge(i) => {
            if ignore_space_before_cursor && i > 0 { Some(words[i - 1]) } else { None }
        },
        CursorPosition::OnWordRightEdge(i) => Some(words[i]),
    }
}
```

## Completion Display

### print_completion_list()

```rust
fn print_completion_list(
    completions: &[String],
    highlighted: Option<usize>,
    output_buf: &mut String,
) -> io::Result<usize> {
    let (w, _) = termion::terminal_size()?;
    let max_word_size = completions.iter().fold(1, |m, x| max(m, x.chars().count()));
    let cols = max(1, w as usize / (max_word_size));
    let col_width = 2 + w as usize / cols;
    let cols = max(1, w as usize / col_width);

    for (index, com) in completions.iter().enumerate() {
        if i == cols { output_buf.push_str("\r\n"); i = 0; }
        if Some(index) == highlighted {
            // Black text on white background
            write!(output_buf, "{}{}", color::Black.fg_str(), color::White.bg_str());
        }
        write!(output_buf, "{:<1$}", com, col_width);
        if Some(index) == highlighted {
            write!(output_buf, "{}{}", color::Reset.bg_str(), color::Reset.fg_str());
        }
        i += 1;
    }
}
```

**Display characteristics:**

- Multi-column layout based on terminal width
- Column count = terminal width / max word size
- Left-aligned with padding to column width
- Selected completion highlighted with **black on white**
- No description support — completions are plain strings
- No color/style support for individual completions

### Comparison with Other Shells

| Feature | Ion (Liner) | Bash (Readline) | Fish | Zsh |
|---------|-------------|-----------------|------|-----|
| Display format | Multi-column | Multi-column | Pager with descriptions | Multi-column with descriptions |
| Highlighting | Black on white | Reverse video | Selected item highlight | Selected item highlight |
| Descriptions | Not supported | COMP_TYPE=63 only | Always shown | Right column |
| Cycling | Tab cycles one at a time | Menu-complete option | Pager navigation | Menu-select |
| Colors | Not supported | Not supported | Full color | Full color via zstyle |

## Autosuggestions

Ion supports fish-like autosuggestions from history:

### How It Works

1. As the user types, `current_autosuggestion()` searches history for the newest entry that starts with the current buffer
2. The suggestion is displayed in **yellow** after the cursor
3. The suggestion text does not replace the current input

```rust
fn current_autosuggestion(&mut self) -> Option<Buffer> {
    if self.show_autosuggestions {
        self.cur_history_loc
            .map(|i| &context_history[i])
            .or_else(|| {
                context_history
                    .get_newest_match(Some(context_history.len()), &self.new_buf)
                    .map(|i| &context_history[i])
            })
    } else {
        None
    }
}
```

### Accepting Suggestions

| Key | Action |
|-----|--------|
| `Ctrl+F` | Accept autosuggestion |
| `Right Arrow` (at end of line) | Accept autosuggestion |

```rust
pub fn accept_autosuggestion(&mut self) -> io::Result<()> {
    if self.show_autosuggestions {
        let autosuggestion = self.autosuggestion.clone();
        let buf = self.current_buffer_mut();
        match autosuggestion {
            Some(ref x) => buf.insert_from_buffer(x),
            None => (),
        }
    }
    self.move_cursor_to_end_of_line()
}
```

## History System

### The History Struct

```rust
pub struct History {
    pub buffers: VecDeque<Buffer>,
    pub append_duplicate_entries: bool,
    pub inc_append: bool,        // Append each entry immediately
    pub share: bool,             // Share across shells
    pub file_size: u64,
    pub load_duplicates: bool,
}
```

### History Search

- **Ctrl+R** — incremental reverse search through history
- **Up/Down arrows** — navigate history entries
- **Autosuggestions** — show matching history as you type

### History Configuration in Ion

```sh
history +inc_append     # Save each command immediately
history +shared         # Share history between shells
history -duplicates     # Don't save duplicate entries
```

### HISTORY_IGNORE

```sh
export HISTORY_IGNORE = [
    'no_such_command'     # exact match
    'whitespace'          # commands starting with space
    'duplicates'          # duplicate entries
    regex:'^cd '          # regex pattern
]
```

## The Buffer Type

`Buffer` represents the current line being edited:

- Supports character-level and grapheme-level operations
- Tracks cursor position
- Provides methods for insertion, deletion, and movement
- Used by both the editor and history system

## Event System

The Liner event system allows completers to react to editor events:

```rust
pub struct Event<'a, 'out: 'a, W: Write + 'a> {
    pub editor: &'a mut Editor<'out, W>,
    pub kind: EventKind,
}

pub enum EventKind {
    BeforeKey(Key),      // Before keypress handling
    AfterKey(Key),       // After keypress handling
    BeforeComplete,      // Before completion processing
}
```

Ion's `IonCompleter` uses `BeforeComplete` to determine the `CompletionType` based on cursor position. This is the only event ion handles.

## Comparison with Readline (Bash)

| Feature | Liner (Ion) | Readline (Bash) |
|---------|-------------|-----------------|
| Library | redox_liner (Rust) | GNU Readline (C) |
| Config file | None (set in initrc) | `.inputrc` |
| Key binding syntax | Rust code only | `bind` builtin + `.inputrc` |
| Macro support | None | `macro` in `.inputrc` |
| Completion variables | None | `COMP_*` variables |
| Completion display | Multi-column, cycling | Multi-column, menu-complete |
| Autosuggestions | Built-in (from history) | Not built-in (requires bash-preexec) |
| Description in completions | Not supported | COMP_TYPE=63 only |
| Custom word breaks | `word_divider_fn` | `COMP_WORDBREAKS` |
| Bell style | Not configurable | `bell-style` in `.inputrc` |
| Bracketed paste | Not supported | Supported |

## References

- [redox_liner crate documentation](https://docs.rs/redox_liner/0.5.1/liner/)
- [redox_liner Editor struct](https://docs.rs/redox_liner/0.5.1/liner/struct.Editor.html)
- [redox_liner Context struct](https://docs.rs/redox_liner/0.5.1/liner/struct.Context.html)
- [redox_liner Completer trait](https://docs.rs/redox_liner/0.5.1/liner/trait.Completer.html)
- [redox_liner KeyMap trait](https://docs.rs/redox_liner/0.5.1/liner/trait.KeyMap.html)
- [redox_liner History struct](https://docs.rs/redox_liner/0.5.1/liner/struct.History.html)
- [redox_liner Emacs keymap](https://docs.rs/redox_liner/0.5.1/liner/struct.Emacs.html)
- [redox_liner Vi keymap](https://docs.rs/redox_liner/0.5.1/liner/struct.Vi.html)
- [Ion completer.rs source](https://github.com/redox-os/ion/blob/master/src/binary/completer.rs)

## Related Skills

- [references/completion.md](references/completion.md) — how IonCompleter uses the Completer trait
- [references/startup-config.md](references/startup-config.md) — keybindings and history configuration
- **bash** skill → `references/readline.md` — comparison with GNU Readline
