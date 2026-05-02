# Papertrading — AI context

## Git commit messages

Same as root `.cursorrules` (SCM / generate-commit reads this file reliably).

### Mandatory

1. **Emoji on line 1:** `type(optional-scope): <emoji><subject>` — Unicode emoji from `package.json` → `config.cz-emoji.types` for that type. One space after `:`; **no** space between emoji and subject. Never `feat:` without emoji.

2. **Body line length:** If you include a body (after a blank line), **every body line must be ≤ 100 characters**. Commitlint rejects **101+**. Count characters per line yourself—do not output one long wrapped paragraph. Split into multiple physical lines. **Re-check each body line before finishing.**

### Types / emoji

Use `name` + `emoji` from `config.cz-emoji.types` only.

Full checklist: `.cursor/rules/conventional-commits.mdc`.
