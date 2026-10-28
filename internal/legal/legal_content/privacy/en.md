# Smart Sewa — User Privacy Policy

**Version:** 1.0.0-draft.1
**Effective date:** 2026-10-27
**Authoritative language:** English
**Data controller:** Sunrise Ceylon Holdings (Pvt) Ltd
**Data Protection Officer:** info@smartsewa.com
**Internal reference:** LP-02-USER-PP

> **Pre-launch draft (v2).** Core company details filled. Remaining `[TBD before launch]` markers (registration number, external processor names, DPA contact) still require values before public release. Final text awaits review by qualified Sri Lankan counsel.

---

## 1. About this policy

This Privacy Policy explains how `Sunrise Ceylon Holdings (Pvt) Ltd` ("**we**") collects, uses, shares, and safeguards personal data of Users of the `Smart Sewa` Platform.

We are the **data controller** for User personal data. Our Data Protection Officer can be contacted at `info@smartsewa.com`.

This policy is written to comply with Sri Lanka's **Personal Data Protection Act, No. 9 of 2022** ("**PDPA**"). Where a jurisdiction with different or stricter requirements applies to a specific User, we will honour the stricter requirement.

## 2. What personal data we collect

The categories below correspond to the Consent & Data Register §3.

**Provided by you:**
- **Identity:** mobile number, password (hashed), full legal name (if provided), date of birth, country, email (optional).
- **Profile:** profile photo (optional), saved addresses, address notes.
- **Content:** work-note titles, descriptions, photos, chat messages, reviews.
- **Recovery:** backup email and security question (optional).
- **Payment (v2):** payment provider token (not card details).

**Collected automatically:**
- **Location:** precise GPS at work-note publish, saved addresses; not tracked continuously.
- **Device:** device ID, OS + app version, push notification token, IP address (transient).
- **Usage:** app events (screens visited, actions taken) with hashed IDs — see BA §12.
- **Crash / performance logs:** no personal content.

**From third parties:**
- **From payment provider (v2):** transaction confirmations, refund events.
- **From reviews left by Workers about you:** your rating.

## 3. Why we collect it — purposes and lawful basis

Per Consent Register §4:

| Purpose | Data used | Lawful basis (PDPA) |
|---|---|---|
| Account operation (login, notifications, matching, hire) | Identity, device, location | Contract |
| Age verification | Date of birth | Legal obligation |
| Fraud / abuse prevention | Device, IP, GPS check, behavioural | Legitimate interest |
| Support and dispute resolution | Chat, call metadata, job history | Contract, legitimate interest |
| Regulatory record-keeping (v2 payments) | Transaction records | Legal obligation |
| Marketing communications | Contact + preferences | Consent |
| Non-essential analytics | Hashed events | Consent |
| Product analytics (essential) | Hashed events | Legitimate interest |
| Recovery via Support | Backup email, security question | Consent |

## 4. Who we share your data with

Per Consent Register §6:

- **A hired Worker** receives your full name (first name minimum), address, address note, and access to a masked call channel — only after you confirm the hire.
- **Bidding Workers who are not yet hired** see only your approximate area, the work-note details, and (via chat, if you shortlist them) messages you send.
- **Our support and trust-and-safety staff** access your data as necessary to resolve tickets and disputes, under audit.
- **Third-party processors.** We use the following main processors, all under written data processing agreements:
  - **Google Cloud Platform** — infrastructure hosting (preferred regions: Singapore or Sri Lanka where offered).
  - **Cloudflare** — edge protection, WAF, DDoS, Turnstile (US / global edges).
  - **`[TBD before launch]`** — SMS OTP and critical notifications.
  - **WhatsApp Business API (Meta)** — WhatsApp OTP fallback and critical notifications.
  - **Firebase Cloud Messaging (Google) and Apple Push Notification service** — push notification delivery.
  - **`[TBD before launch]`** (v2 only) — payment processing and payouts.
  - **`[TBD before launch]`** — product analytics (hashed IDs only).
  - **`[TBD before launch]`** — crash and error monitoring (no PII).
  A current list of all processors is available on request from our DPO and will be updated as it changes.
- **Law enforcement and regulators** where legally required, upon receipt of a valid legal request.
- **Optional partners** (insurance, financial services in v4) only with your explicit consent per offer.

We do **not** sell your personal data.

## 5. Cross-border transfers

Some of the processors listed in Section 4 (notably Cloudflare, WhatsApp/Meta, Firebase Cloud Messaging, and Apple Push Notification service) are located outside Sri Lanka.

Where we transfer personal data outside Sri Lanka, we rely on one or more of the following lawful bases, consistent with the **Personal Data Protection Act, No. 9 of 2022**:

- **Contract necessity** — where the transfer is required to deliver the Platform to you.
- **Explicit consent** — obtained via this Privacy Policy's disclosure and, where relevant, via specific consent screens in the app.
- **Contractual safeguards** — every processor has agreed in writing to security, access control, breach notification, sub-processor management, and audit obligations at least equivalent to those required of us as controller.
- **Provider certifications** — we prefer processors that hold recognised certifications (for example, ISO 27001, SOC 2) as an additional safeguard.

Where the Data Protection Authority publishes adequacy decisions or approved contractual arrangements, we will update our transfers to rely on those mechanisms and update this Policy accordingly.

**Data residency preference.** For services offering a choice of region, we prefer processing in Sri Lanka or Singapore. A current data-residency map is available from our DPO.

## 6. How long we keep your data

Per Consent Register §5. Summary:

- Account data: for the duration of your account, plus 7-day deletion grace period.
- Chat messages: 90 days after job closure, or until any related dispute is closed.
- Work-note content: 12 months after job closure.
- Reviews: retained (anonymised on account deletion).
- Transaction records (v2): 7 years (tax law).
- Device / IP session data: 90 days.
- Consent records: retained indefinitely as evidence of consent.
- Analytics event data: 24 months raw; aggregated indefinitely.

## 7. How we protect your data

We implement technical and organisational measures aligned with **OWASP MASVS / ASVS**, **CIS Google Cloud Foundations Benchmark**, and other industry standards described in our Security Standards. Key controls include encryption in transit (TLS 1.3) and at rest (with customer-managed keys for sensitive data), strict access controls with full audit logging, and multi-factor authentication for staff.

No system is perfectly secure. If we ever suffer a personal data breach that presents a real risk to you, we will notify you and the Data Protection Authority in accordance with the PDPA (targeted within 72 hours of confirmation).

## 8. Your rights under the PDPA

Per Consent Register §8 you have the following rights:

- **Right of access** — obtain a copy of your personal data.
- **Right to rectification** — correct inaccurate data.
- **Right to erasure** — request deletion (subject to legal retention obligations).
- **Right to restriction** — freeze processing pending review.
- **Right to data portability** — receive your data in a machine-readable format.
- **Right to object** — object to specific processing.
- **Rights regarding automated decision-making** — receive a meaningful explanation and request human review.
- **Right to withdraw consent** where consent is the basis.
- **Right to lodge a complaint** with the Data Protection Authority.

To exercise any right, use Settings → Privacy → **My data** in the app, or contact `info@smartsewa.com`. We aim to respond within 7 days; the statutory maximum is 30 days.

**Fees.** The first request in respect of any right is free. For repeated or manifestly excessive requests we may charge a reasonable fee reflecting the administrative cost, as permitted by law; we will tell you in advance if this applies.

**Identity verification.** Before serving any DSR request, we will verify your identity to protect against impersonation. For higher-sensitivity requests (KYC access, account deletion), additional verification steps may be required.

## 9. Automated decision-making

Per Consent Register §12 we use automated systems to:
- Match your work-notes to Workers (based on distance, availability, service type, rating).
- Order bidders on your list (rating × distance × experience).
- Flag unusual patterns (fraud, safety) for human review.

Material decisions that affect you significantly (account suspension, ban, refund, dispute outcome) are always **made by a human**, not by an automated system. You can request a meaningful explanation of any automated processing by contacting the DPO.

## 10. Children

The Platform is for adults **18 or older**. We do not knowingly collect personal data from anyone under 18. If we learn that a person under 18 has provided data to us, we will delete it.

## 11. Cookies (web only, from v1.5)

Please see the Cookie Policy (Doc 6).

## 12. Changes to this policy

We may update this policy from time to time in our sole discretion. **Material changes** (for example, changes to what personal data we collect, or changes to your data subject rights under this Policy) will be notified in-app at least **7 days before they take effect** (unless a shorter period is reasonable in the circumstances or required by law), accompanied by a plain-language summary, and — where required by law — will require your re-acceptance in accordance with UC-15 before you continue using the Platform. **Non-material changes** are versioned and dated and take effect on posting. Your **continued use of the Platform after a change takes effect constitutes acceptance** of it.

## 13. Contact

Data Protection Officer: `info@smartsewa.com`
Data Protection Authority (Sri Lanka): `[TBD before launch]`
