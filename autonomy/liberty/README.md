---
status: draft
tags: [autonomy]
builds_on: [foundations/ethics]
constrains: [justice/enforcement]
related: [health/preventive, law/statutory, autonomy/rights, health/mental]
---

# Liberty

Individual freedom to make choices about one's own body, mind, and life, provided one is mentally capable of informed self-direction.

## Substance Access

Most substances, including those currently illegal, should be legal for competent adults. Prohibition fails: it criminalizes personal choice, funds black markets, and prevents quality control.

Conditions for legal access:
- The person is not mentally unwell to the degree they cannot make coherent decisions
- The person is not being coerced
- The person can demonstrate informed self-direction

### Evidence-Based Substance Tiers

Tier classification is based on actual harm data, not legacy legislation. Alcohol has far more documented deaths and negative effects than cannabis, so it sits in a higher tier. All classifications are subject to revision as evidence accumulates.

| Tier | Criteria | Waiver rigor | Examples (pending objective study) |
|------|----------|-------------|-----------------------------------|
| 1 - Low risk | Low addiction, low lethality, low social harm | Light waiver, basic ID | Cannabis, psilocybin |
| 2 - Moderate | Moderate addiction/lethality, documented health degradation | Full waiver, health baseline, credit tracking | Alcohol, nicotine, MDMA |
| 3 - High | High addiction, significant health damage, overdose risk | Full waiver, mandatory narcan access, health monitoring, tighter credit limits | Methamphetamine, heroin, cocaine |
| Prohibited | Potency-to-amount ratio makes weaponization trivial | Illegal | Fentanyl, carfentanil |

The prohibited tier is not moralistic. If micrograms kill and the substance is easily dispersed, it is a chemical weapon, not a drug. That is a security concern (see [[security/threat-response]]), not a liberty one.

### Waiver System

Every substance, including alcohol and nicotine, requires a signed waiver. The waiver contains:
- Valid ID verification
- Acknowledgment of known side effects
- Binding commitment to narcan availability (where applicable)
- Agreement to access treatment resources if desired (detox, counseling)
- Agreement to not operate machinery while intoxicated
- Agreement to public intoxication restrictions

Waivers are digitally signed, RFC 3161 timestamped, and backed up to paper. See [[security/resilience]] for the digital infrastructure and [[law/statutory]] for implementation.

### Credit System

A usage tracking system monitors frequency and quantity per person per substance:

1. Normal usage patterns - no intervention
2. Escalating patterns - health check-in triggered (not forced, but recorded)
3. Addict-level usage - treatment access activated, health tracking mandatory, access to that substance rate-limited
4. Recovery - credits normalize over time as usage decreases

This is stewardship in action. You are not banned. When the data shows you are hurting yourself, the system increases its involvement proportionally. The guiding hand gets firmer as the situation worsens.

Drugs that degrade health should limit your access to that drug as degradation progresses. The credit system makes this automatic and evidence-based, not arbitrary.

### Public Intoxication

The health effects of drugs on psyche, mental health, and physical health are real. Legality does not mean endorsement. Your altered state affects others in shared spaces, so public intoxication remains prohibited.

## Bodily Autonomy

People have the right to modify their own bodies: tattoos, piercings, cosmetic surgery, hormone therapy, gender-affirming procedures.

The constraint: the person must be making a healthy, informed choice, not acting from a place of mental illness or external pressure. See [[autonomy/identity]] for gender and body modification specifics, and [[health/mental]] for how competency is assessed without becoming gatekeeping.

The line is not "we disapprove of your choice" but "you are not currently capable of making this choice well."
