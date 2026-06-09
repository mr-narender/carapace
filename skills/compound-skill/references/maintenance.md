# Maintenance

How to keep a compound skill accurate and up to date as the project evolves.

## Update Triggers

A compound skill should be updated when:

| Trigger | What to Update |
|---------|---------------|
| **New feature added** | Add or update the relevant reference document; add keywords to the Sub-Resources table; add Quick Guide entries |
| **Behavior changed** | Update the affected reference document; check if cross-references are still accurate |
| **API renamed** | Update type/function names in keywords, reference content, and Quick Guide |
| **Bug fixed** | Remove or update any "known issue" or "gotcha" that described the bug |
| **New sub-topic identified** | Create a new reference document; add it to the Sub-Resources table and Quick Guide |
| **Sub-topic removed** | Remove the reference document; update the Sub-Resources table and Quick Guide; check for links to the removed document |
| **Related skill created or updated** | Update Cross-Project References; check for overlap that can now be delegated |

## Adding a Reference Document

1. Create `references/new-topic.md` following the structure in [reference-doc.md](reference-doc.md)
2. Add a row to the Sub-Resources table in SKILL.md with keywords and the link
3. Add relevant entries to the Quick Guide
4. Check if any existing reference documents should link to the new one
5. Check if the new document overlaps with an existing skill — if so, add a cross-reference

## Removing a Reference Document

1. Delete `references/old-topic.md`
2. Remove its row from the Sub-Resources table
3. Remove its Quick Guide entries
4. Search all other reference documents for links to the removed document and update them
5. Check if the Cross-Project References section needs updating

## Keeping Current

### Source Code Changes

The most reliable way to keep a skill current is to verify against the source code. When reviewing a skill:

1. Check that code examples still compile or match the current source
2. Verify that type names and function signatures are current
3. Confirm that control flow descriptions match the current implementation
4. Test any command examples against the current version

### Version Skew

Skills can drift from the source as the project evolves. Mitigate this by:

- **Attributing sources** — include "Source of truth" links so the agent knows where to verify
- **Dating sections** — for rapidly evolving areas, note the version or date the information was verified
- **Flagging unstable APIs** — mark experimental or unstable features so the agent knows to verify them
- **Checking on release** — review the skill when a new version of the project is released

### Stale Indicators

Watch for these signs that a reference document is stale:

- Code examples that don't match the current source
- Type or function names that no longer exist
- References to removed features or deprecated APIs
- "Known issues" that have been fixed
- Links to other skills that no longer exist or have been reorganized

## Refactoring a Skill

As a skill grows, it may need restructuring:

### Splitting a Reference

When a reference document exceeds ~300 lines or covers two distinct concerns:

1. Identify the natural split point (e.g., "general concept" vs. "project-specific integration")
2. Create the new reference document
3. Move the relevant content, adding cross-links between the two
4. Update the Sub-Resources table with two rows
5. Update the Quick Guide

### Merging References

When two references are tightly coupled and always loaded together:

1. Merge the smaller into the larger
2. Add a redirect note at the top of the removed file (or just delete it)
3. Update the Sub-Resources table
4. Update all links in other reference documents

### Reorganizing the Routing Table

When the Sub-Resources table grows large:

- Group related rows by adding a brief category header before them
- Ensure the Quick Guide still maps to the correct documents
- Verify that keyword coverage is still complete (no gaps in routing)
