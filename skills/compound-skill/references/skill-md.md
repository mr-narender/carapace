# SKILL.md — The Routing Table

How to write the SKILL.md file for a compound skill. The SKILL.md is the entry point and routing table — it tells the AI agent which reference document to load for a given task.

## Frontmatter

Every SKILL.md begins with YAML frontmatter:

```yaml
---
name: my-skill
description: >
  Use when working with [project/topic] — [key areas covered].
  Triggers on: "keyword1", "keyword2", "specific term", "AnotherTerm",
  "compound term", "TypeA", "TypeB.method", "CONFIG_VAR".
user-invocable: true
---
```

### `name`

A short, lowercase, hyphenated identifier. Must be unique across all skills the agent can access.

### `description`

A multi-line string that serves two purposes:

1. **Trigger definition** — the agent uses this to decide when to activate the skill. Include:
   - The project/topic name and common aliases
   - Key type names, function names, and configuration variables
   - Domain-specific terms a user might mention
   - Phrases in quotes that should trigger the skill

2. **Scope summary** — briefly state what the skill covers so the agent can also decide when *not* to activate it

The `Triggers on:` line lists the exact strings and terms that should cause the agent to load this skill. Be comprehensive — include technical terms, API names, config variables, and common abbreviations.

### `user-invocable`

Set to `true`. This allows the skill to be loaded on demand.

## Title and Introduction

After the frontmatter:

```markdown
# [Skill Name] In-Depth Reference

One-sentence summary of what the skill covers, with a link to the project.
```

Keep the introduction to 1–2 sentences. The detail lives in the reference documents.

## Sub-Resources Table

The core of the routing table. A Markdown table mapping **keywords** to **reference documents**:

```markdown
## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

|| Keywords | Reference ||
||----------|----------|
|| keyword1, keyword2, TypeA, methodB, concept name | [references/topic-a.md](references/topic-a.md) |
|| keyword3, keyword4, CONFIG_VAR, edge case description | [references/topic-b.md](references/topic-b.md) |
```

### Keywords Column

List the terms, type names, function names, and concepts that the reference document covers. These keywords serve as the agent's lookup index — when a user's query mentions one of these terms, the agent knows which reference to load.

Guidelines:

- **Include type names**: `Action`, `Model`, `Server`, `Config`
- **Include function/method names**: `Handle()`, `Process()`, `Invoke()`
- **Include concept names**: `dispatch engine`, `middleware chain`, `event loop`
- **Include config variables**: `DEBUG`, `LOG_LEVEL`, `CONFIG_PATH`
- **Include domain jargon**: `revset`, `fileset`, `compspec`, `middleware`
- **Comma-separated**: list all relevant keywords for that document
- **Be comprehensive**: better to over-list than under-list — the agent uses these to route

### Reference Column

A Markdown link to the reference file. Always use relative paths: `[references/topic.md](references/topic.md)`.

## Quick Guide

A question→reference mapping for common tasks. This section helps the agent answer "which doc do I load for X?":

```markdown
## Quick Guide

- **How do I [common task]?** → [references/topic-a.md](references/topic-a.md)
- **How does [concept] work?** → [references/topic-b.md](references/topic-b.md)
- **What is [thing]?** → [references/topic-c.md](references/topic-c.md)
- **How do I debug [problem]?** → [references/topic-a.md](references/topic-a.md) and [references/topic-b.md](references/topic-b.md)
```

Guidelines:

- **Phrase as questions** a developer would actually ask
- **One question per line**, with an arrow (`→`) to the reference
- **Multiple references** are fine when a question spans topics
- **Cover the most common tasks** — this is a quick lookup, not an exhaustive list
- **Don't repeat the keywords table** — the Quick Guide answers "how do I..." while the Sub-Resources table answers "where is X documented?"

## Data Flow Diagram (Optional)

For skills covering a system with a clear pipeline or dispatch flow, include a data flow diagram before the Sub-Resources table:

```markdown
## Data Flow

\`\`\`
Input event
  → parser
    → middleware chain
      → handler dispatch
        → response formatting
          → output
\`\`\`
```

This gives the agent a mental model of the system before it dives into specific references. Only include this when the topic has a clear sequential flow — skip it for reference-style topics (e.g., a language reference, a theme format reference).

## Cross-Project References

Link to other skills that cover related but distinct topics. This section prevents overlap by delegating:

```markdown
## Cross-Project References

- For [related topic A], use the **other-skill** skill (in the other-repo repo).
- For [related topic B], use the **another-skill** skill → `references/specific-doc.md`.
- For [related topic C] internals, see the **third-skill** skill.
```

Guidelines:

- **Link, don't duplicate** — if another skill already covers a topic, reference it rather than re-documenting
- **Be specific** — when referencing a sub-section of another skill, include the specific reference path
- **Explain the boundary** — briefly state what the other skill covers that this one doesn't
- **List all overlapping skills** — the agent needs to know when to switch

## Real-World Examples

These existing compound skills demonstrate the patterns described here:

| Skill | Type | References | Notable Feature |
|-------|------|------------|----------------|
| **carapace-dev** | Library development | 20+ | Large routing table, per-shell sub-references |
| **bash** | Shell internals | 5 | Per-concern decomposition |
| **jj** | VCS reference | 12 | Concept-heavy, CLI reference |
