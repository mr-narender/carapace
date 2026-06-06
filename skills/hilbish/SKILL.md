---
name: hilbish
description: >
  Use when working with hilbish shell internals — completion system, hilbish.completions,
  completion handler, completion groups (grid, list), readline integration, Lua API,
  bait event system, runner mode, commander, lunacolors, syntax highlighting,
  hinter, vim mode, job control, or init.lua configuration. Triggers on: "hilbish",
  "hilbish completion", "hilbish.completions", "hilbish.completion.handler",
  "completion group", "grid completion", "list completion", "hilbish editor",
  "hilbish.runner", "runner mode", "hybrid runner", "bait.catch", "commander",
  "lunacolors", "hilbish.highlighter", "hilbish.hinter", "hilbish.prompt",
  "hilbish.alias", "hilbish.inputMode", "hilbish.vim", "GopherLua",
  "hilbish.editor", "init.lua", "TabCompleter", "CompletionGroup".
user-invocable: true
---

# Hilbish Shell In-Depth Reference

Comprehensive reference for hilbish shell internals, with emphasis on the completion system and how external tools hook into it.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

| Keywords | Reference |
|----------|----------|
| hilbish.completions, hilbish.completion.handler, completion group, grid, list, TabCompleter, CompletionGroup, hcmpAdd, hcmpBins, hcmpFiles, hcmpDirs, hcmpCall, hcmpHandler, binaryComplete, fileComplete, dirComplete, matchPath, escapeFilename, charEscapeMap, splitForFile, luaCompletions, scope, command.<cmd>, query, ctx, fields, prefix, items, descriptions, displays, aliases, Suggestions, TrimSlash, NoSpace, TabDisplayGrid, TabDisplayList, carapace hilbish integration, overriding completion handler, sudo completion, cd completion | [references/completion.md](references/completion.md) |
| readline, maxlandon/readline, lineReader, rl.go, TabCompleter, HintText, SyntaxHighlighter, ViModeCallback, ViActionCallback, Searcher, fuzzy search, hilbish.editor, readline.new, deleteByAmount, getLine, getVimRegister, insert, log, read, getChar, setVimRegister, emacs mode, vim mode, multiline prompt, SetPrompt, SetRightPrompt, RefreshPromptInPlace, fileHistory, Ctrl-R history search | [references/line-editing.md](references/line-editing.md) |
| Lua, GopherLua, golua, runner mode, hybrid, hybridRev, sh, lua, hilbish.runner, hilbish.runnerMode, hilbish.run, commander, commander.register, Sink, bait, bait.catch, bait.catchOnce, bait.hooks, bait.release, bait.throw, lunacolors, lunacolors.format, hilbish.alias, hilbish.appendPath, hilbish.which, hilbish.goro, hilbish.interval, hilbish.timeout, hilbish.read, hilbish.exec, hilbish.cwd, hilbish.highlighter, hilbish.hinter, hilbish.prompt, hilbish.multiprompt, hilbish.inputMode, event-driven, hooks, signals, yarn, multi-threading | [references/language.md](references/language.md) |
| execution, job control, job handler, job struct, job.start, job.done, job.add, jobs.all, jobs.get, jobs.disown, jobs.last, foreground, background, stop, start, wait, disown, SIGHUP, exitCode, cmd, running, pid, stdout, stderr, runInput, handleLua, splitInput, finishExec, continuePrompt, command.preexec, command.exit, command.not-found, command.not-executable, processors, aliases.resolve | [references/execution.md](references/execution.md) |
| init.lua, XDG, hilbish.userDir, hilbish.dataDir, .hilbishrc.lua, startup, nature/, nature/init.lua, nature/completions/, nature/commands/, nature/processors/, nature/editor.lua, nature/hilbish.lua, nature/vim.lua, nature/runner.lua, hilbish.opts, autocd, history, greeting, motd, fuzzy, notifyJobFinish, processorSkipList, SHLVL, package.path, module.load, start/, prompt verbs, %d, %u, %h, fmtPrompt, configuration, sample config | [references/startup-config.md](references/startup-config.md) |

## Quick Guide

- **How does hilbish's completion system work?** → [references/completion.md](references/completion.md)
- **How does the readline integration handle tab completion?** → [references/line-editing.md](references/line-editing.md)
- **What is the Lua API and runner mode system?** → [references/language.md](references/language.md)
- **How does hilbish handle job control and signals?** → [references/execution.md](references/execution.md)
- **Which startup files does hilbish read?** → [references/startup-config.md](references/startup-config.md)
- **How do I override the completion handler?** → [references/completion.md](references/completion.md)
- **How do I register a command completer?** → [references/completion.md](references/completion.md)
- **What are completion groups (grid vs list)?** → [references/completion.md](references/completion.md)
- **How do I enable vim mode?** → [references/line-editing.md](references/line-editing.md)
- **How do I customize the prompt?** → [references/startup-config.md](references/startup-config.md)
- **What is the bait event system?** → [references/language.md](references/language.md)
- **How does runner mode work?** → [references/language.md](references/language.md)
- **How does carapace integrate with hilbish?** → [references/completion.md](references/completion.md)

## Cross-Project References

For carapace-specific hilbish integration (snippet, value formatting), see the **carapace-dev** skill → `references/shell.md`.
