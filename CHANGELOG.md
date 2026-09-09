# Changelog

All notable changes to this project are documented in this file.

## v0.1.4 (2026-09-09)

Full changelog: https://github.com/jonbaldie/tui-reader/compare/v0.1.3...v0.1.4

### Bug Fixes

- **Highlight only the selected link instance (#58, fixes #55).** Link selection styling compared only the selected link's target, so every attached link on the page sharing that target was rendered with `selectedLinkStyle`. Selection is now styled by page-local link index (`Model.selectedLink`) while scanning attached markup in document order, so only the active instance is highlighted.
- **Keep links with prefix punctuation intact when wrapping (#59).** Links whose opening bracket `[` is preceded by punctuation without whitespace — e.g. `([Target Heading](#target-heading))` — were torn apart across display lines when the source line wrapped, breaking markdown rendering, link attachment, and Tab navigation. `wrapTokens` now keeps such link markup intact.

### Documentation

- **Configure GitHub issues as the agent issue tracker (#57).** Added agent configuration for issue tracking, triage labels, and domain docs.