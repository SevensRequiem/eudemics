---
status: draft
tags: [security]
builds_on: [decision-making/delegation, autonomy/privacy]
related: [infrastructure/communication, law/constitutional, law/codification]
---

# Resilience

How the system survives disasters, attacks, and infrastructure failure.

## Digital Infrastructure with Physical Backup

Everything is digital for efficiency. No paper bureaucracy for daily operations. But everything gets routinely:
- Printed to paper
- Backed up to redundant drives
- Replicated across nodes in the global data pool
- Encrypted at rest

Three layers of redundancy: digital (fast), physical drives (medium-term), paper (survives EMP, fire, floods if stored properly).

## RFC 3161 Timestamping

All audit logs, governance decisions, legal records, citizen data events, and git commits use RFC 3161 Trusted Timestamping Authority. Cryptographic proof of when things happened. This makes records tamper-evident and legally defensible.

## Global Data Replication

Each node contributes to a global pool of data that is replicated everywhere:
- No single node holds unique critical data
- If a node is destroyed, its data survives in every other node
- Citizen data packs, governance records, scientific knowledge, cultural archives: all replicated
- This is the backbone of war resilience: lose a node, lose nothing but the physical infrastructure

## Survival Stores

Every node maintains emergency reserves:
- Water supply for extended self-sufficiency
- Shelf-stable/infinite-shelf-life food
- Simple electronics and machines sufficient to reboot local society
- Medical supplies
- Communication equipment (independent of main infrastructure)

These are not aspirational. They are maintained, inventoried, and tested regularly.

## State Archiving

Archiving and data redundancy is constitutional law (see [[law/constitutional]]). Wayback Machine equivalents receive state funding or are state-absorbed.

The complete record of human knowledge and digital activity is a public good. No corporation, no state action, no legal dispute can erase the archive. The archive is inviolable. This is not just policy, it is foundational law.

## Inter-Node Communication

- EMP-hardened and highly redundant
- Multiple physical media: microwave relays, fiber optic, satellite
- Multiple routes to and from every node (no single point of failure)
- TOR-like relay routing as the default: node traffic routes through intermediary nodes, not direct peer-to-peer
- Relay routing creates automatic audit trail on every relay node (metadata, not content)
- Prevents covert node-to-node communication that bypasses the system
- Inter-node encryption with post-quantum algorithms (ML-KEM-768 for key exchange, ML-DSA-65 for signatures)

See [[infrastructure/communication]] for the physical infrastructure details.

## Node Override Keys

State write-access override keys (for death, incapacitation, guardianship) require multi-node authorization. No single node can forge records or override citizen data sovereignty. See [[autonomy/privacy]] for the full data pack architecture.
