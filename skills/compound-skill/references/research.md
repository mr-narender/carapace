# Research Phase

How to research a project or topic before writing a compound skill. The research phase determines what to cover, at what depth, and how to decompose the topic into reference documents.

## Overview

A compound skill is only as good as its research. Before writing any reference document, invest time in understanding the project's architecture, mental model, and the questions a developer will actually ask. The goal is to produce a set of reference documents that cover the topic comprehensively without overlap.

## Information Sources

Research from multiple angles to build a complete picture:

| Source | What It Provides | How to Use |
|--------|-----------------|------------|
| **Source code** | Ground truth — actual behavior, edge cases, internal APIs | Read the core types, the main entry points, and the dispatch/control flow. Trace the happy path and error paths. |
| **Official documentation** | Intended behavior, user-facing API, configuration | Read the full docs. Note where docs diverge from code (document both). |
| **Tutorials and guides** | Common workflows, mental models, "how to think about it" | Extract the conceptual model — what are the key abstractions? |
| **Blog posts and talks** | Design rationale, historical context, non-obvious gotchas | Capture the "why" behind design decisions. |
| **Issue trackers** | Known bugs, edge cases, user confusion points | Document workarounds and common pitfalls. |
| **Test suites** | Expected behavior, edge cases, invariants | Tests reveal what the code *should* do, including cases docs omit. |

## Topic Decomposition

### Identify the Core Abstractions

Start by identifying the main types, concepts, or subsystems. For a library, these are typically:

- **Core types** — the primary data structures and their relationships
- **Control flow** — how a request flows through the system
- **Configuration** — how behavior is customized
- **Extension points** — how users extend or hook into the system
- **Edge cases** — non-obvious behavior, gotchas, known issues

For a tool or application:

- **Commands and flags** — the CLI surface
- **Conceptual model** — the mental model users need
- **Configuration** — settings, profiles, environment variables
- **Integration** — how it connects to other tools
- **Workflows** — common task sequences

### Determine Sub-Topics

Group related concepts into sub-topics that can each become a reference document. Apply these principles:

1. **One concern per document** — each reference covers a single coherent topic (e.g., "completion system", "template language", "theme format")
2. **Self-contained but not isolated** — a reader should understand the topic from the document alone, with links to related topics for deeper context
3. **Right granularity** — not too broad (one giant doc) and not too narrow (one doc per function). If a topic needs more than ~300 lines, consider splitting it
4. **Name by noun, not question** — use `completion.md` not `how-completion-works.md`. The routing table handles the question→document mapping

### Scope Boundaries

Define what the skill covers and — critically — what it does **not** cover:

- **In scope**: internals, APIs, configuration, edge cases, cross-cutting concerns specific to the topic
- **Out of scope**: topics covered by other skills (link to them instead), general programming knowledge, unrelated tooling

When a topic is partially covered by another skill, document only the aspect unique to your project and link to the other skill for the rest. For example, a carapace shell integration document covers *how carapace formats output for bash*, but links to the **bash** skill for *how bash's completion system works internally*.

## Audience Analysis

Write for a developer who:

- Is working with the project's codebase (not a first-time user)
- Needs to understand internals, not just the public API
- Will ask "how does X work?" and "why does X behave this way?"
- May need to debug, extend, or modify the system

This means: include internal types, show code paths, explain design rationale, and document edge cases. Don't repeat the user guide — provide the depth that source code alone doesn't convey.

## Research Checklist

Before writing, confirm:

- [ ] You've read the core source files (not just the public API)
- [ ] You've traced the main control flow end-to-end
- [ ] You've identified the key types and their relationships
- [ ] You've read the official documentation for accuracy
- [ ] You've checked tests for edge cases and invariants
- [ ] You've identified which topics belong in this skill vs. other skills
- [ ] You've defined the sub-topic list and their scope boundaries
- [ ] You've checked for existing skills that overlap (to link, not duplicate)
