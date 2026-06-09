---
name: compound-skill
description: >
  Use when creating a compound skill — a structured knowledge base for a project or topic
  that uses a SKILL.md routing table and reference documents. Covers the research phase,
  topic decomposition, SKILL.md frontmatter and routing table format, reference document
  structure, cross-referencing, and maintenance. Triggers on: "compound skill", "create skill",
  "write skill", "skill template", "skill structure", "routing table", "reference document",
  "skill authoring", "skill creation guide".
user-invocable: true
---

# Creating a Compound Skill

Guide for creating a compound skill — a structured knowledge base that provides in-depth reference documentation for a project, technology, or topic. A compound skill consists of a **SKILL.md** routing table and a set of **reference documents** in a `references/` directory.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

|| Keywords | Reference |
||----------|----------|
|| research, source code analysis, documentation review, topic decomposition, sub-topic identification, scope boundaries, audience analysis, information sources | [references/research.md](references/research.md) |
|| SKILL.md, frontmatter, name, description, triggers, user-invocable, routing table, Sub-Resources table, Quick Guide, Cross-Project References, data flow diagram | [references/skill-md.md](references/skill-md.md) |
|| reference document, structure, heading, code example, table, cross-reference, source attribution, depth level, naming convention | [references/reference-doc.md](references/reference-doc.md) |
|| cross-reference, link, overlap avoidance, delegation, Cross-Project References, stay on topic, DRY, complementary scope | [references/cross-referencing.md](references/cross-referencing.md) |
|| maintenance, update triggers, adding references, removing references, keeping current, version skew, changelog | [references/maintenance.md](references/maintenance.md) |

## Quick Guide

- **How do I start creating a compound skill?** → [references/research.md](references/research.md)
- **How do I write the SKILL.md routing table?** → [references/skill-md.md](references/skill-md.md)
- **How do I structure a reference document?** → [references/reference-doc.md](references/reference-doc.md)
- **How do I handle cross-references without duplicating?** → [references/cross-referencing.md](references/cross-referencing.md)
- **How do I keep a skill up to date?** → [references/maintenance.md](references/maintenance.md)
- **What goes in the frontmatter description?** → [references/skill-md.md](references/skill-md.md)
- **How do I decide on sub-topics?** → [references/research.md](references/research.md)
- **How do I name reference files?** → [references/reference-doc.md](references/reference-doc.md)
- **When should I link to another skill instead of writing?** → [references/cross-referencing.md](references/cross-referencing.md)

## Cross-Project References

- For the carapace library's own compound skill structure, see the **carapace-dev** skill (in this repo).
- For shell-specific compound skills (bash, zsh, fish, etc.), see the corresponding shell skills in this repo's `skills/` directory.
- For the tabdance compound skill, see the **tabdance** skill (in the tabdance-bin repo).
- For the jj compound skill, see the **jj** skill (in the carapace-jjlex repo).
