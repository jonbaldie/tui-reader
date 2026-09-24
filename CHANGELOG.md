# Changelog

All notable changes to this project are documented in this file.

## v0.1.8 (2026-09-24)

Full changelog: https://github.com/jonbaldie/tui-reader/compare/v0.1.7...v0.1.8

### Bug Fixes

- **Own default page geometry (#126).** `Book` now owns normalized page dimensions, including the 60-column by 20-line default, and dump mode pads pages using the effective page height.
- **Align fence scanners with indented code (#130, fixes #129).** Anchor and link scanners now treat four-space-indented fence-looking lines as indented code before changing fence state, matching the formatter and keeping real headings and links navigable.

## v0.1.7 (2026-09-21)

Full changelog: https://github.com/jonbaldie/tui-reader/compare/v0.1.6...v0.1.7

### Bug Fixes

- **Reflow consecutive lines in prose as single paragraphs (#84, fixes #77).** Consecutive nonblank Markdown source lines are now reflowed and soft-wrapped as one paragraph instead of being paginated line by line.
- **Adapt layout to compact dimensions and long titles (#86).** Regressions for width/height clipping in compact terminals and wrapping of long titles.
- **Ignore link markup inside inline code spans (#85, fixes #78).** Links written inside inline code were extracted as selectable page links; inline-code spans no longer contribute links.
- **Ignore inline code when styling links (#90).** Inline code spans matching page links were styled as active links and stole Tab selection.
- **Track link index by page link position (#87, fixes #82).** Link selection is tracked by page-local position so over-wide links stay selectable and navigable.
- **Keep over-wide links on their display page (#88, fixes #83).** Links wider than the page no longer jump to a later page when selected or followed.
- **Keep repeated internal links on their display pages (#95, fixes #89).** Repeated internal links with the same target now resolve to the page they appear on instead of the first occurrence.
- **Anchor page 0 to first content and pages to first heading (#97, fixes #91).** The initial `WindowSizeMsg` reflow anchored to the last heading on the default page, skipping page 1 on startup.
- **Truncate footer page info in compact terminals (#98, fixes #92).** The footer status line no longer overflows or clips in narrow terminals.
- **Keep provenance mapping in range when wrap misses (#102, fixes #99).** Wrapped provenance mapping clamps to the page so rendering cannot panic on a missed wrap point.
- **Keep boundary links on their display line (#103, fixes #100).** A stray bracket at a reflowed source-line boundary no longer swallows the real link.
- **Avoid indented overflow for wide runes (#104, fixes #101).** Paragraphs containing wide runes rewrap at the full page width rather than overflowing the two-column indent.
- **Avoid indented overflow for wide runes in code blocks (#107, fixes #105).** Indented code blocks with wide runes fall back to full-width wrapping so display lines never exceed the page width.
- **Preserve links spanning physical lines in soft-wrapped paragraphs (#110, fixes #106).** Links split across source-line boundaries in prose paragraphs are attached and navigable again.
- **Keep non-ASCII letters in anchors and normalize link fragments (#116, fixes #112).** Anchors keep letters in any script, and link fragments are normalized before lookup, so `#第一章` and `#été` resolve.
- **Skip fenced code when extracting headings and reflowing (#120, fixes #111).** Headings and links inside fenced code blocks are no longer treated as navigable Markdown, and fence contents are not joined as prose.
- **Strip UTF-8 BOM so first heading is recognised (#119, fixes #113).** A leading UTF-8 byte-order mark no longer hides the first heading and its anchor.
- **Keep plain-text source lines from merging into paragraphs (#124, fixes #114).** Non-Markdown files render one source line per display line instead of collapsing into wrapped paragraphs.

## v0.1.6 (2026-09-11)

Full changelog: https://github.com/jonbaldie/tui-reader/compare/v0.1.5...v0.1.6

### Bug Fixes

- **Ignore indented code in navigation extraction (#70, fixes #67).** Anchor and source-link extraction scanned raw lines independently of the formatter's four-space code-line rule, so headings and links inside indented code blocks were exposed as navigable Markdown. The indented-code predicate is now shared across formatting, extraction, position anchoring, and TUI heading styling.
- **Keep adjacent indented code lines together (#71, fixes #68).** The formatter inserted an inter-paragraph spacer before every nonblank raw line after the first, including continuations of an indented code block. Only code-block continuations now suppress the spacer; paragraph transitions retain it.
- **Reject malformed `--dump=N` values (#72, fixes #69).** Malformed `--dump=N` values were silently accepted and changed dump scope; they are now rejected.

### CI

- **Bump pinned Go to 1.26.8 for mutago v2.10.7 (#76, fixes #75).** mutago v2.10.7 requires Go >= 1.26.6, so the pinned 1.26.5 failed the mutation workflow's install step.

### Documentation

- **Cap local Mutago runs at one worker (#74, fixes #73).**
- **Preserve 2026-09-10 exploratory testing report (#80).**

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