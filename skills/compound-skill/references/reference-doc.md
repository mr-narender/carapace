# Reference Document Structure

How to write individual reference documents — the files in the `references/` directory that contain the actual knowledge.

## File Naming

Use short, lowercase, hyphenated names that describe the topic as a noun:

| Good | Bad | Reason |
|------|-----|--------|
| `completion.md` | `how-completion-works.md` | Name by topic, not question |
| `action.md` | `actions.md` | Singular for a cohesive concept |
| `shell-bash.md` | `bash-shell-formatting.md` | Follow a consistent prefix pattern |
| `build-release.md` | `building-and-releasing.md` | Concise |
| `v1-v2-migration.md` | `migrating-from-v1-to-v2.md` | Concise |

When a skill has multiple documents in a category, use a consistent prefix:

- `shell.md`, `shell-bash.md`, `shell-zsh.md`, `shell-fish.md` — per-variant references
- `completion.md`, `editor.md`, `startup-config.md` — per-concern references

## Document Structure

Each reference document follows this general structure (adapt to the topic):

```markdown
# [Topic Title]

One-sentence summary of what this document covers.

## [First Major Section]

Content with code examples, tables, and explanations.

## [Next Major Section]

...

## Edge Cases and Known Issues  (if applicable)

...

## References  (if applicable)

- Link to source files, official docs, or related references

## Related Skills  (if applicable)

- Links to other skills or other reference documents in this skill
```

### Title

Start with an H1 (`#`) that names the topic. Use a descriptive noun phrase:

- `# Bash Programmable Completion`
- `# The Action Type and Modifiers`
- `# The Request Pipeline`
- `# Core Concepts`

### Opening Summary

One sentence stating what the document covers. This helps the agent decide if this is the right document to load.

### Source Attribution

When the document is derived from specific source material, attribute it at the top:

```markdown
> **Source of truth**: <https://example.com/docs/topic>. For **related topic**, see [related.md](related.md).
```

This tells the agent where to verify information and which other reference to consult for adjacent topics.

## Content Guidelines

### Code Examples

Show real code from the project. Use fenced code blocks with language tags:

````markdown
```go
func (a Action) Invoke(ctx Context) InvokedAction {
    // ...
}
```
````

Guidelines:

- **Show actual code**, not pseudocode — the agent needs to understand the real implementation
- **Include context** — show enough surrounding code that the reader understands where this fits
- **Annotate sparingly** — use inline comments only for non-obvious behavior, not to narrate line-by-line
- **Show the happy path first**, then edge cases

### Tables

Use Markdown tables for structured information — type references, comparison tables, configuration options:

```markdown
| Field | Type | Description |
|-------|------|-------------|
| `Value` | `string` | The completion candidate value |
| `Display` | `string` | How the value appears in the menu |
| `Description` | `string` | Short description shown alongside |
```

### Diagrams

Use ASCII art for flow diagrams and state machines:

```markdown
```
Input → parse() → validate() → handler.Dispatch() → format()
```
```

Only include diagrams when they add clarity — not every document needs one.

### Depth Level

Write for a developer who is **working with the codebase**, not a first-time user. This means:

- **Include internal types** — not just the public API
- **Show code paths** — trace through the implementation
- **Explain design rationale** — why, not just what
- **Document edge cases** — the non-obvious behavior that trips people up
- **Include gotchas** — thread safety, global state, ordering dependencies

Don't:

- Repeat the user guide or README
- Document every exported function (that's what godoc/pkg.go.dev is for)
- Include installation instructions (unless they're non-obvious)

### Cross-References Within a Skill

Link to other reference documents in the same skill when the topic overlaps:

```markdown
For how the dispatch engine routes requests, see [dispatch.md](dispatch.md).
```

Use relative links within the skill: `[other.md](other.md)`, not absolute paths.

### Cross-References to Other Skills

When a topic is covered by another skill, link to that skill rather than re-documenting:

```markdown
For the internal dispatch system (routing, middleware, error handling), see the **framework** skill → `references/dispatch.md`.
```

This is the key mechanism for avoiding overlap. See [cross-referencing.md](cross-referencing.md) for detailed guidelines.

## Section Order Convention

While not every document needs every section, follow this order when applicable:

1. **Title and summary** — what this document covers
2. **Source attribution** — where the information comes from (if applicable)
3. **Overview / Core concept** — the mental model
4. **Main sections** — the topic's sub-areas
5. **Edge cases and known issues** — non-obvious behavior
6. **References** — links to source files, official docs
7. **Related skills** — links to other skills or references
