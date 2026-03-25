---
status: draft
tags: [infrastructure]
builds_on: [security/resilience, decision-making/delegation]
related: [autonomy/privacy, knowledge/information-flow]
---

# Communication

Physical and digital communication infrastructure.

## Node-Internal

Each node is responsible for its own internal communication infrastructure. Evidence-based investment: nodes assess their needs and build accordingly.

## Inter-Node

EMP-hardened and highly redundant:
- Multiple physical media per route: microwave relays, fiber optic, satellite
- Multiple routes between every pair of nodes (no single point of failure)
- TOR-like relay routing as the default, not direct peer-to-peer
- Relay routing creates automatic audit trail on intermediary nodes (metadata, not content)
- Prevents covert node-to-node communication bypassing the system
- Post-quantum encryption: ML-KEM-768 for key exchange, ML-DSA-65 for signatures

## Cross-Node Assistance

Any node can assist another with communication infrastructure if their schedule is clear and they can spare resources. Evidence-based prioritization.

## Resilience

The communication network must survive EMP, natural disaster, and targeted attack. Redundancy across media types (microwave survives what fiber does not, and vice versa) ensures no single failure mode takes down inter-node communication.
