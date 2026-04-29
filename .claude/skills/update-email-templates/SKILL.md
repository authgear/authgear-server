---
name: update-email-templates
description: Update Authgear email templates using the correct source files, translation files, and commit order. Use when editing email wording, email structure, or subject lines.
---

# Update Email Templates

Follow this workflow whenever updating email templates.

## Scope Rules

1. Only edit source template files:
   - `resources/authgear/templates/en/messages/*.txt.gotemplate`
   - `resources/authgear/templates/en/messages/*.mjml.gotemplate`
2. Do not edit generated template files directly:
   - `resources/authgear/templates/en/messages/*.txt`
   - `resources/authgear/templates/en/messages/*.mjml`
3. In `*.txt.gotemplate` and `*.mjml.gotemplate`, use translations from:
   - `resources/authgear/templates/en/messages/translation.json`
4. Always edit English templates only at source stage.
   - Do not manually update non-`en` locales in source-edit commits.
5. Email subject translations are defined in:
   - `resources/authgear/templates/en/translation.json`

## Workflow

1. Edit source templates first:
   - `resources/authgear/templates/en/messages/*.txt.gotemplate`
   - `resources/authgear/templates/en/messages/*.mjml.gotemplate`
   - `resources/authgear/templates/en/messages/translation.json` (body/content translations)
   - `resources/authgear/templates/en/translation.json` (subject translations)
2. Generate `.mjml` from `.mjml.gotemplate`:
   - `go run ./scripts/generatemjml/main.go -i resources/authgear/templates`
3. Commit source edits and generated `.mjml` in one commit.
4. In a separate commit, generate non-English translations:
   - `make -C scripts/python generate-translations`
   - `ANTHROPIC_API_KEY` must be set.
   - If `ANTHROPIC_API_KEY` is not set, stop and ask the user to run this step.
5. Finally, generate HTML emails:
   - `make html-email`

## Commit Boundaries (Required)

- Commit 1:
  - Source edits (`*.gotemplate`, `en/messages/translation.json`, optional `en/translation.json`)
  - Generated `.mjml`
- Commit 2:
  - Generated non-English translations (`make -C scripts/python generate-translations`)
- Commit 3:
  - Generated email HTML artifacts (`make html-email`)

Do not combine generated artifacts (`.html`, `.txt`, non-`en` translations) in the same commit as source `*.gotemplate` edits.
