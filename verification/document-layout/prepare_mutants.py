from pathlib import Path
import shutil
import sys

source = Path(__file__).resolve().parent
out = Path(sys.argv[1]).resolve()
for name in ("BlockContext", "LayoutCache", "WidthAccumulator", "FenceContext", "SourcePositions"):
    target = out / name
    target.mkdir(parents=True, exist_ok=True)
    shutil.copy2(source / (name + ".tla"), target)
    config = (source / (name + ".cfg")).read_text().replace("Mutate = FALSE", "Mutate = TRUE")
    (target / (name + ".cfg")).write_text(config)
mutations = {
    "token": ("emittedSpace' = (kind # \"space\" /\\ pending)", "emittedSpace' = TRUE"),
    "indent": ("accepted' = fits", "accepted' = TRUE"),
    "provenance": ("IF cursor < 3 THEN cursor+1 ELSE cursor", "IF cursor < 3 THEN cursor+1 ELSE cursor+1"),
    "columns": ("accepted' = (column' >= 4)", "accepted' = TRUE"),
    "digits": ("accepted' = (digits /\\ ch = \"suffix\")", "accepted' = TRUE"),
    "title": ("titled' = previousSpace", "titled' = ~previousSpace"),
    "lines": ('(ch = "LF" /\\ ~wasLF)', '(ch = "LF")'),
}
for name, (old, new) in mutations.items():
    target = out / ("scanner-" + name)
    target.mkdir(parents=True, exist_ok=True)
    text = (source / "ScannerDecisions.tla").read_text()
    assert old in text, old
    (target / "ScannerDecisions.tla").write_text(text.replace(old, new))
    shutil.copy2(source / "ScannerDecisions.cfg", target)
extras = {
    "SourcePositions": [("explicit", "ELSE expected", "ELSE fallback")],
    "PageAnchors": [("priority", 'ELSE IF firstHeading >= 0 THEN firstHeading ELSE IF firstContent >= 0 THEN firstContent ELSE carry', 'ELSE carry')],
    "ProsePromotion": [("adjacency", 'collected > 0 /\\ kind = "underline"', 'kind = "underline"')],
    "PositionLookup": [("best", "best' = IF ok THEN", "best' = IF FALSE THEN")],
    "RegistryFolds": [("range", "k <= end /\\ rawPages[k]", "k < end /\\ rawPages[k]"), ("last-heading", '[headings EXCEPT ![key] = Len(history)]', 'IF headings[key] = -1 THEN [headings EXCEPT ![key] = Len(history)] ELSE headings'),
                      ("lookup", "answer' = rawPages[query']", "answer' = -1")],
}
for module, changes in extras.items():
    for label, old, new in changes:
        target = out / (module + "-" + label)
        target.mkdir(parents=True, exist_ok=True)
        text = (source / (module + ".tla")).read_text()
        assert old in text, old
        (target / (module + ".tla")).write_text(text.replace(old, new, 1))
        shutil.copy2(source / (module + ".cfg"), target)
print(out)
