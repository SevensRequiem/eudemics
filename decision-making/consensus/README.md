---
status: draft
tags: [decision-making]
builds_on: [decision-making/delegation]
related: [security/resilience, resource-systems/accounting]
---

# Consensus

How nodes make collective decisions and coordinate.

## Multi-Node Voting

Decisions affecting multiple nodes require multi-node consensus, similar to a VRF (Verifiable Random Function) committee.

A single node governs its own local affairs. When a decision affects the network, a committee of randomly selected nodes deliberates and votes. The VRF selection prevents collusion - you cannot predict or manipulate which nodes will be on the committee.

## Node Denial Override

A node CAN deny a request for resources or assistance. But if the requesting node and other nodes vote that the help is needed, the denying node is overridden. The system serves the whole.

This is not tyranny of the majority. The override requires clear evidence that the denial harms the network. A node that legitimately cannot spare resources without harming itself is not overridden.

## Emergency Coordination

- The affected node is the primary responder for local emergencies
- Neighboring nodes assist if they can without hurting themselves
- Global nodes check resource pools and offer help and manpower for large-scale events
- Automated signals with node keys enable instant notification and resource checking
- The system is interconnected: if one node hurts, the system hurts. Helping a struggling node is not charity, it is self-maintenance. Strongest-as-weakest-link principle.

### Per-Node Preparedness

Each node knows its local environment and prepares accordingly:
- Coastal nodes prepare for flooding, tsunamis
- Fault-line nodes prepare for earthquakes
- Every node maintains survival stores: water, shelf-stable food, simple electronics and machines sufficient to reboot local society

### Global Unknowns

Preparation for low-probability high-impact events: disease outbreak, asteroid impact, sea level rise, climate change, alien contact. "No evidence of aliens" does not mean zero preparation for contact is practical. Even limited preparation for unlikely events is sound engineering.

## Political Parties

There are none. The system is git-based and VRF-based governance. Policy is proposed, debated on evidence, and voted on by competent nodes. There are no parties, no campaigns, no political messaging. See [[law/creation]] for the legislative process.
