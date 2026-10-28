# legal_content/

Canonical source for the Smart Sewa user-facing legal documents (Terms &
Conditions, Privacy Policy) in English, Sinhala, and Tamil. Both the Flutter
mobile app and the Next.js web app read from here so the exact bytes shown to
the user — and therefore the hash recorded in the consent ledger — are
identical across platforms.

**Current version:** `1.0.0-draft.1` · effective `2026-10-27` · **pre-launch**.

## Files

```
legal_content/
├── config.json            Versions, paths, SHA-256(version + "\n" + body) per language
├── terms/
│   ├── en.md              Authoritative text (User T&C)
│   ├── si.md              Sinhala translation
│   └── ta.md              Tamil translation
└── privacy/
    ├── en.md              Authoritative text (User Privacy Policy)
    ├── si.md              Sinhala translation
    └── ta.md              Tamil translation
```

**English is authoritative.** Sinhala and Tamil are translations for user
convenience. If a translation drifts from the English version, English
prevails (stated inside each document).

## How the version string works

- `1.0.0-draft.1` — pre-launch wording. Must be replaced with `1.0.0` once the
  `[TBD before launch]` placeholders are filled and Sri Lankan counsel has
  reviewed. Bumping the version salts the SHA-256 displayed-text hash, which
  forces re-consent of every existing user. That is the intended behaviour.
- Non-material changes (typos, formatting): bump the trailing identifier
  (`1.0.0 → 1.0.1`). Existing consents remain valid under PDPA because the
  material terms are unchanged — but their recorded hash will no longer match
  the live text, so a re-consent pass is still the safest default.
- Material changes (anything that changes rights, obligations, data flows, or
  retention): bump the minor version (`1.0.0 → 1.1.0`) and require re-consent.
  See UC-15 and `CONSENT_REGISTER.md` R-17 to R-19.

## Regenerating config.json

Any time a `.md` file in this folder is edited (even whitespace), the hashes
in `config.json` must be regenerated. From the repo root:

```sh
python3 scripts/update_legal_hashes.py   # TODO: not yet created; see Task #7
```

Until the helper exists, the Python block in the second commit of this
feature can be copied out of git history.

## What still needs to be filled

Grep for `[TBD before launch]` in this folder. Expected hits: registration
number, registered office address, DPO email, support email, DPA contact.
Any change to these values is a material change under §19 and forces
re-consent.

## Known legal gaps (not fixed in code)

- **Lawyer review** — the full drafts have not been reviewed by qualified
  Sri Lankan counsel yet. Marked `pre_launch: true` in `config.json`.
- **Worker-side docs (Doc 3 / Doc 4)** — not in this folder. Will land when
  the worker app is built. See `LEGAL_REVIEW.md` blocker #2.
- **Translator notes** — the Sinhala/Tamil Terms carry a note that §17 has
  been revised in English and must be re-translated before launch.

See `CONSENT_REGISTER.md` for the full compliance picture.
