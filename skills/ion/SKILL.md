---
name: ion
description: >
  Use when working with ion shell internals — completion system, Liner library,
  Completer trait, IonCompleter, IonFileCompleter, MultiCompleter, CursorPosition,
  quoting/expansion, execution model, startup files, initrc, keybindings, prompt
  hooks, or ion-plugins. Triggers on: "ion", "ion shell", "ion completion",
  "IonCompleter", "IonFileCompleter", "MultiCompleter", "Completer trait",
  "redox_liner", "liner", "initrc", "ion-plugins", "CompletionType",
  "CursorPosition", "keybindings vi", "keybindings emacs", "fn PROMPT",
  "^>", "^|", "&>", "&|", "let", "sigil", "$sigil", "@sigil".
user-invocable: true
---

# Ion Shell In-Depth Reference

Comprehensive reference for ion shell internals, with emphasis on the completion system and how external tools hook into it.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

|| Keywords | Reference |
||----------|----------|
| IonCompleter, IonFileCompleter, MultiCompleter, Completer trait, CompletionType, CursorPosition, BeforeComplete event, filename_completion, escape/unescape, command completion, variable completion, file completion, PATH scanning, tab completion, completion cycling, completion display, nospace, carapace ion integration, JSON suggestion format | [references/completion.md](references/completion.md) |
| Liner, redox_liner, Editor, Context, KeyMap, KeyBindings, emacs mode, vi mode, key bindings, Tab, Ctrl+F, autosuggestion, history, Buffer, word divider, get_buffer_words, print_completion_list, complete() method, read_line() | [references/line-editing.md](references/line-editing.md) |
| ion syntax, variables, arrays, let, sigils, $, @, type system, string methods, array methods, slicing, brace expansion, command substitution, process substitution, quoting, single quotes, double quotes, escape sequences, functions, fn, control flow, if, for, while, matches, test, redirection, ^>, ^|, &>, &|, pipelines, glob | [references/language.md](references/language.md) |
| execution, subshell, background jobs, fg, bg, jobs, disown, wait, signals, exit status, &&, ||, &, exec, eval, source, suspend, and, or, not | [references/execution.md](references/execution.md) |
| initrc, XDG_CONFIG_HOME, startup files, keybindings, set -o, history configuration, HISTFILE, HISTORY_IGNORE, PROMPT function, color namespace, env namespace, aliases, environment variables, ion-plugins | [references/startup-config.md](references/startup-config.md) |

## Quick Guide

- **How does ion's completion system work?** → [references/completion.md](references/completion.md)
- **How does the Liner library handle tab completion?** → [references/line-editing.md](references/line-editing.md)
- **What are the ion language syntax and quoting rules?** → [references/language.md](references/language.md)
- **How does ion handle job control and signals?** → [references/execution.md](references/execution.md)
- **Which startup files does ion read?** → [references/startup-config.md](references/startup-config.md)
- **How does carapace integrate with ion?** → [references/completion.md](references/completion.md)
- **Can I add custom completions to ion?** → [references/completion.md](references/completion.md)
- **How do I configure vi/emacs keybindings?** → [references/startup-config.md](references/startup-config.md)
- **How do I customize the prompt?** → [references/startup-config.md](references/startup-config.md)
- **What are the $ and @ sigils?** → [references/language.md](references/language.md)
- **How does stderr piping (^|, ^>) work?** → [references/language.md](references/language.md)

## Cross-Project References

For carapace-specific ion integration (snippet, value formatting), see the **carapace-dev** skill → `references/shell.md` (ion is listed as a secondary shell).
