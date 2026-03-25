---
status: draft
tags: [autonomy]
builds_on: [foundations/ethics, security/resilience]
constrains: [decision-making/accountability]
related: [knowledge/information-flow, law/statutory]
---

# Privacy

Sovereign ownership of personal data. The individual controls their information, not the state.

## Personal Data Packs

Every citizen has an encrypted data pack containing all their personal data. The architecture:

- **Dual-held:** you hold the primary copy, the state holds a backup
- **Write access:** state can only write through YOUR keys. The state records events (substance credits, health events, legal records) but needs your cryptographic authorization to do so.
- **Keys:** one-time use, post-quantum (ML-DSA-65 for signatures, ML-KEM-768 for encryption). Each write operation uses a fresh key.
- **You control access:** you grant the state access to specific data with TTL (time-to-live) or permanent grants, with revocation at any time.
- **Dual-signature verification:** systems verify via both state AND personal data signatures. The state cannot spoof your data because it needs your signature. You cannot deny state-recorded events because they carry the state's signature.

### Physical ID Backup

The data pack has a physical card form - an encrypted ID card that bridges from legacy identity systems. Everyone understands an ID card. The card contains enough to bootstrap your digital identity if the digital system is unavailable. Post-quantum encrypted, biometric-linked, replaceable through the multi-signee master key process if lost.

### State Override Keys

Override keys exist for death, incapacitation, and guardianship:
- Every override use is fully auditable, RFC 3161 timestamped, logged to immutable records
- Override requires multi-node authorization (not a single official)
- Override scope is limited to what is needed (death: records transfer; guardianship: relevant medical/financial data)
- See [[decision-making/delegation]] for the federated node structure that prevents single-point abuse

### Data and Features

If you withhold data from the state, you receive "basic" access to features that require that data. Data enables curated solutions - less data means less personalization. This is a trade-off you make freely, not a punishment.

Some state functions require data. Others do not. You choose what to share, understanding the consequences.

### Internet Identity

- Every internet user has a KYC-linked ID through the data pack system
- Sites NEVER see your actual identity
- Your ISP holds a hash of your ID key
- The state can match the hash to your real identity only when legally required (crime investigation), with full audit trail
- Competency tier keys are separate sub-keys - sites verify your tier without knowing who you are
- This eliminates: LLM bot farms, sockpuppets, coordinated inauthentic behavior, most spam, and many fraud vectors
- Privacy is preserved for legitimate use. Accountability is enforced for abuse.

### Master Key Architecture

- Multiple signees required to create an ID master key (prevents single point of compromise)
- Master key generates sub-keys for specific purposes: internet tier access, substance waiver/credit system, health records, financial records
- Sub-keys can be revoked and regenerated independently
- If the master key is compromised, the multi-signee process creates a new one

### State Tracking

The state needs to track certain metadata for crime correlation and system operation. This is acknowledged openly, not hidden. The scope of what the state tracks by default is defined in law, auditable, and subject to revision. See [[law/statutory]] for the legal framework.
