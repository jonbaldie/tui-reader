# Changelog

All notable changes to this project are documented in this file.

## v0.1.5 (2026-09-10)

Full changelog: https://github.com/jonbaldie/tui-reader/compare/v0.1.4...v0.1.5

### Bug Fixes

- **Preserve reading and history positions across resize (#64, fixes #61).** On terminal resize, the reader clamped the current page and back-history entries to the new page count, losing the reader's place and truncating history. Each page is now anchored to a representative raw source line and page indices are remapped around `Reflow` through those anchors, so reading position and history survive re-pagination.
- **Wrap error messages to terminal width (#65, fixes #62).** The centered error view rendered `Error: <err>` as one unwrapped line, so at narrow terminal widths the terminal clipped the filename and OS reason. The error style now sets its width to the terminal width so lipgloss wraps the message.

## v0.1.4 (2026-09-09)

Full changelog: https://github.com/jonbaldie/tui-reader/compare/v0.1.3...v0.1.4

### Bug Fixes

- **Highlight only the selected link instance (#58, fixes #55).** Link selection styling compared only the selected link's target, so every attached link on the page sharing that target was rendered with `selectedLinkStyle`. Selection is now styled by page-local link index (`Model.selectedLink`) while scanning attached markup in document order, so only the active instance is highlighted.
- **Keep links with prefix punctuation intact when wrapping (#59).** Links whose opening bracket `[` is preceded by punctuation without whitespace — e.g. `([Target Heading](#target-heading))` — were torn apart across display lines when the source line wrapped, breaking markdown rendering, link attachment, and Tab navigation. `wrapTokens` now keeps such link markup intact.

### Documentation

- **Configure GitHub issues as the agent issue tracker (#57).** Added agent configuration for issue tracking, triage labels, and domain docs.