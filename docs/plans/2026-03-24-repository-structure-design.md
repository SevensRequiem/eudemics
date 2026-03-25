---
status: approved
date: 2026-03-24
---

# Repository Structure Design

## Context

Eudemics is a conceptual governance model optimized toward the human condition. It needs a structure that is deeply modular, never requires restructuring, and supports continuous refinement. Git history serves as the audit trail for how the model evolves.

## Decisions

- **Content format**: Markdown with YAML frontmatter for metadata (tags, relationships, status) and inline [[wiki links]] for casual cross-references
- **Structure**: Functional decomposition into independent packages, inspired by Go module design. Each package represents a distinct functional domain that can be reasoned about independently.
- **Cross-references**: Combination of all approaches - inline wiki links, "related" sections, and typed frontmatter relationships (constrains, builds-on, implements, related)
- **Critique**: Integrated per-package rather than centralized. Each package can contain analysis of how existing systems handle that function.
- **Viewer**: Future SPA (single index.html with inline CSS/JS) that reads a generated manifest to render the wiki with navigation and graph view.

## Package Structure

16 top-level packages organized by function, not by traditional political categories:

| Package | Scope |
|---------|-------|
| foundations | Axioms: human needs, human nature, ethics |
| autonomy | Individual freedoms, rights, privacy, identity |
| decision-making | How groups make choices: consensus, delegation, participation, accountability |
| resource-systems | Material needs: accounting, distribution, production, ownership |
| law | Rule lifecycle: creation, amendment, constitutional, statutory, codification |
| justice | When things break: conflict resolution, enforcement, restoration |
| knowledge | How society learns: education, information flow, research |
| community | How people relate: belonging, culture, support |
| health | Physical, mental, preventive, emergency |
| ecology | Environmental constraints: sustainability, resource limits, stewardship |
| infrastructure | Physical/digital systems: communication, housing, mobility |
| security | Protection: defense, resilience, threat response |
| relations | Inter-community: diplomacy, migration, trade, borders |
| technology | Tech governance: regulation, automation, ethics, deployment |
| transition | (Deferred) How to get from here to there |

## Frontmatter Schema

```yaml
---
status: draft | review | stable | deferred
tags: [parent-package, ...]
builds-on: [concept-path, ...]
constrains: [concept-path, ...]
implements: [concept-path, ...]
related: [concept-path, ...]
---
```

## Future Work

- SPA viewer with manifest generation
- Graph visualization of concept relationships
- Contribution guidelines for the model itself
