#!/usr/bin/env python3
from pathlib import Path
import argparse
import shutil
import subprocess
import tempfile

here = Path(__file__).resolve().parent
repo = here.parents[1]
parser = argparse.ArgumentParser()
parser.add_argument('--mutation', action='store_true')
args = parser.parse_args()
with tempfile.TemporaryDirectory(prefix='reader-layout-') as scratch:
    target = Path(scratch)
    for name in ('go.mod', 'go.sum'):
        shutil.copy2(repo / name, target / name)
    package = target / 'internal/book'
    package.mkdir(parents=True)
    for source in (repo / 'internal/book').glob('*.go'):
        if not source.name.endswith('_test.go'):
            shutil.copy2(source, package / source.name)
    shutil.copy2(here / 'code_test.go.txt', package / 'formal_layout_test.go')
    def run():
        return subprocess.run(['go', 'test', './internal/book', '-count=1'], cwd=target, text=True, capture_output=True)
    baseline = run()
    print(baseline.stdout, end='')
    print(baseline.stderr, end='')
    if baseline.returncode:
        raise SystemExit(baseline.returncode)
    if args.mutation:
        source = package / 'book.go'
        original = source.read_text()
        needle = 'prevCode, prevList = false, false'
        assert original.count(needle) == 1
        source.write_text(original.replace(needle, needle + '\n\t\t\tfirstParagraph = true'))
        mutated = run()
        print(mutated.stdout, end='')
        print(mutated.stderr, end='')
        assert mutated.returncode != 0 and 'first-block privilege reset' in mutated.stdout, 'Mutation escaped the behavioral check'
        source.write_text(original)
        restored = run()
        print(restored.stdout, end='')
        assert restored.returncode == 0, restored.stderr
        print('PASS: real-code mutation caught, isolated source restored, baseline passes.')
