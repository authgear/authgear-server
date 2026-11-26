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
2. Commit the `en` source edits alone. Do not run any generator yet: `generatemjml` rewrites the `.mjml` of every locale, which buries the hand-written change under generated files.
3. Update non-English translations, in a separate commit:
   - **Changed wording:** delete the changed keys from all non-`en` locale files so the generator recreates them, then run it.
     - For `en/messages/translation.json` changes, delete the **top-level key** (e.g. `UsageAlert`, not `UsageAlert.Alert.Title`) from all non-`en` `messages/translation.json` files: `make templates-translation-json-del-key KEY=<TopLevelKey>`
     - For `en/translation.json` subject changes, delete the changed key from all non-`en` `translation.json` files: `make translation-json-del-key KEY=<key>`
     - Run once per changed top-level key, then `make -C scripts/python generate-translations`.
     - Requires being logged in to Claude: `claude login`. If not logged in, stop and ask the user to run `claude login` first.
   - **No wording change** (only a template variable or other non-prose value, e.g. `{{ .URL }}` → `{{ .Link }}`): replace the value directly in every non-`en` file instead, since `generate-translations` re-translates the whole key. Say so in the commit message.
4. Run every generator, and commit only their output:
   - `go run ./scripts/generatemjml/main.go -i resources/authgear/templates` (`.mjml` for every locale)
   - `make html-email` (`.html` and `.txt` for every locale)
   - Check `git status`: only generated `.mjml`, `.html` and `.txt` files may change. Never edit them by hand.

## Commit Boundaries (Required)

Keep each commit reviewable on its own:

- Commit 1: `en` source edits only (`en/messages/*.gotemplate`, `en/messages/translation.json`, optional `en/translation.json`). No generated files.
- Commit 2: non-English translations (step 3).
- Commit 3: generated templates only (`.mjml`, `.html`, `.txt` from step 4), e.g. `Generate <name> templates`.

Generate last: the generated files of every locale read the translations, so they must be built after commit 2. Do not combine generated artifacts with source or translation edits.
