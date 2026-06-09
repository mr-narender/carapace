# Cross-Referencing

How to link between reference documents and other skills without duplicating content. Cross-referencing is the primary mechanism for keeping a compound skill focused and avoiding overlap.

## The Core Principle: Link, Don't Copy

When information is already documented elsewhere, link to it instead of re-stating it. This applies at two levels:

1. **Within a skill** — link to another reference document in the same skill
2. **Across skills** — link to a different skill that covers the topic

## Within a Skill

When one reference document mentions a concept covered in another reference in the same skill, link to it:

```markdown
The traverse engine classifies arguments using pflagfork (see [pflag.md](pflag.md) for flag mode details).
```

Guidelines:

- **Link on first mention** — the first time a concept from another reference appears, link to it
- **Don't link every occurrence** — subsequent mentions don't need links
- **Use relative links** — `[other.md](other.md)`, not absolute paths
- **Brief inline reference** — don't explain the linked topic, just identify it and link

## Across Skills

When a topic is covered by another skill entirely, delegate to that skill:

```markdown
For bash's programmable completion system (COMP_TYPE, COMP_WORDBREAKS, compspec search order, bash-completion helpers), see the **bash** skill → `references/completion.md`.
```

### When to Delegate

Delegate to another skill when:

- The topic is **fundamental to the other project**, not just used by yours
- The other skill already has **comprehensive coverage** of the topic
- Your skill only needs the topic as **background context**, not as a primary subject
- Documenting it in your skill would create a **maintenance burden** (you'd have to keep it in sync with the other project's changes)

### When Not to Delegate

Document the topic in your own skill when:

- Your project has a **specific integration** or **custom behavior** that differs from the general case
- The topic is **internal to your project** (e.g., how your code calls the other project's API)
- The other skill covers the **general concept** but your project has **project-specific edge cases**
- The information is **how-to** for your project specifically, not reference for the other project

### The Partial Overlap Pattern

The most common case is partial overlap: your project uses another project's feature, but adds project-specific behavior on top. Handle this by:

1. **Document the project-specific part** in your reference
2. **Link to the other skill** for the general part
3. **State the boundary explicitly**

Example from the carapace ecosystem:

- The **carapace-dev** skill's `references/shell-bash.md` documents *how carapace formats completions for bash* (project-specific)
- The **bash** skill's `references/completion.md` documents *how bash's completion system works internally* (general)
- Each links to the other at the boundary

## Cross-Project References Section

The SKILL.md's `Cross-Project References` section is the top-level delegation mechanism. It tells the agent which other skills to consult for related topics:

```markdown
## Cross-Project References

- For [topic A], use the **other-skill** skill (in the other-repo repo).
- For [topic B] internals, see the **another-skill** skill → `references/specific-doc.md`.
```

Guidelines:

- **List every overlapping skill** — don't leave the agent guessing
- **Be specific about the boundary** — state what the other skill covers that yours doesn't
- **Include the repo location** — the agent needs to know where to find the skill
- **Reference specific documents** when the overlap is with a sub-section, not the whole skill

## Avoiding Overlap

### The "Stay on Topic" Test

Before adding content to a reference document, ask:

1. **Is this specific to my project?** If yes, document it. If it's general knowledge about another project, link instead.
2. **Would this content need to change if the other project changed?** If yes, it belongs in the other skill.
3. **Am I writing this because I need it, or because it's interesting?** Only document what a developer working with your project actually needs.

### The "Overlap Audit"

After writing a skill, check for overlap:

1. Read each reference document and highlight any content that describes another project's internals
2. For each highlighted section, check if an existing skill covers it
3. If yes, replace the section with a link to that skill
4. If no, consider whether the content belongs in a new skill for that project

### Common Overlap Scenarios

| Scenario | Solution |
|----------|----------|
| Your project uses library X | Document *how you use X*, link to X's skill for *how X works* |
| Your project has a shell integration | Document *your snippet and output format*, link to the shell skill for *the shell's completion internals* |
| Your project extends a framework | Document *your extensions*, link to the framework's skill for *the base framework* |
| Your project converts between formats | Document *your conversion logic*, link to each format's skill for *the format specification* |
