# eudemics

an experiment in my free time to optimize various issues with current ocracies/isms/archies, governments, social constructs, towards the human condition/psyche and not towards any one person, group, or type. this will probably take the resource/energy systems thought of in technocracy as a foundation.

this is a conceptual model, not software. there is no "v1" - it evolves through continuous refinement. git history serves as the audit trail for how the model changes over time.

## structure

each package is a distinct functional domain, decomposed by function rather than traditional political categories. documents use markdown with YAML frontmatter and `[[wiki links]]` for cross-references.

| package | scope |
|---------|-------|
| [foundations](foundations/) | axioms: human needs, human nature, ethics |
| [autonomy](autonomy/) | individual freedoms, rights, privacy, identity |
| [decision-making](decision-making/) | how groups make choices: consensus, delegation, participation, accountability |
| [resource-systems](resource-systems/) | material needs: accounting, distribution, production, ownership |
| [law](law/) | rule lifecycle: creation, amendment, constitutional, statutory, codification |
| [justice](justice/) | when things break: conflict resolution, enforcement, restoration |
| [knowledge](knowledge/) | how society learns: education, information flow, research |
| [community](community/) | how people relate: belonging, culture, support |
| [health](health/) | physical, mental, preventive, emergency |
| [ecology](ecology/) | environmental constraints: sustainability, resource limits, stewardship |
| [infrastructure](infrastructure/) | physical/digital systems: communication, housing, mobility |
| [security](security/) | protection: defense, resilience, threat response |
| [relations](relations/) | inter-community: diplomacy, migration, trade, borders |
| [technology](technology/) | tech governance: regulation, automation, ethics, deployment |
| [transition](transition/) | *(deferred)* how to get from here to there |

## viewer

a self-contained go binary serves a local wiki viewer for browsing the model. build and run:

```
go build -o eudemics .
./eudemics --port 8080
```

flags:
- `--port` - server port (default 8080)
- `--content` - path to content directory (default `.`)
- `--repo` - git repo URL to clone if content dir is empty

the viewer renders markdown server-side, provides a navigable tree, and displays concept relationships in a canvas graph. content changes are hot-reloaded via filesystem watcher.

cross-platform binaries are built on tag push via github actions.

## format

each document follows this structure:

```yaml
---
status: draft | review | stable | deferred
tags: [parent-package, ...]
builds-on: [concept-path, ...]
constrains: [concept-path, ...]
implements: [concept-path, ...]
related: [concept-path, ...]
---

# Title

content with [[wiki links]] for inline cross-references
```
