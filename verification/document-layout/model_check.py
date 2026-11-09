#!/usr/bin/env python3
from pathlib import Path
import os
import shutil
import subprocess
import sys
import tempfile

here = Path(__file__).resolve().parent
artifact = 'http:tla2tools[url=https://github.com/tlaplus/tlaplus/releases/download/v1.7.4/tla2tools.jar,checksum=sha256:936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88]@1.7.4'
java = os.environ.get('JAVA') or subprocess.check_output(['mise', 'which', 'java', '--tool', 'java@temurin-25'], text=True).strip()
jar = os.environ.get('TLA2TOOLS') or str(Path(subprocess.check_output(['mise', 'where', artifact], text=True).strip()) / 'tla2tools.jar')
expected = {
    'BlockContext': 'LaterProseKeepsIndent',
    'LayoutCache': 'LayoutConsistent',
    'WidthAccumulator': 'NoLostColumns',
    'FenceContext': 'CodeHeadingNotRegistered',
    'SourcePositions': 'OffsetsResetAtSourceBoundary',
    'SourcePositions-explicit': 'OffsetsResetAtSourceBoundary',
    'RegistryFolds-range': 'FirstLast',
    'PageAnchors-priority': 'AnchorPriority',
    'ProsePromotion-adjacency': 'AdjacentOnly',
    'RegistryFolds-last-heading': 'FirstLast',
    'RegistryFolds-lookup': 'LookupRetainsFirst',
    'PositionLookup-best': 'GreatestCandidate',
    **{'scanner-' + name: 'ObservedRule' for name in ('token', 'indent', 'provenance', 'columns', 'digits', 'title', 'lines')},
}

def run(spec, work):
    return subprocess.run([java, '-XX:+UseParallelGC', '-cp', jar, 'tlc2.TLC', '-workers', '1', '-config', str(spec.with_suffix('.cfg')), str(spec)], cwd=work, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)

with tempfile.TemporaryDirectory(prefix='reader-layout-models-') as scratch:
    scratch = Path(scratch)
    for spec in sorted(here.glob('*.tla')):
        work = scratch / ('baseline-' + spec.stem)
        work.mkdir()
        copied = work / spec.name
        shutil.copy2(spec, copied)
        shutil.copy2(spec.with_suffix('.cfg'), copied.with_suffix('.cfg'))
        result = run(copied, work)
        assert result.returncode == 0, result.stdout
        print(spec.stem, 'baseline PASS', flush=True)
    mutants = scratch / 'mutants'
    subprocess.run([sys.executable, str(here / 'prepare_mutants.py'), str(mutants)], check=True, stdout=subprocess.DEVNULL)
    assert set(p.name for p in mutants.iterdir()) == set(expected)
    for folder in sorted(mutants.iterdir()):
        spec = next(folder.glob('*.tla'))
        result = run(spec, folder)
        invariant = expected[folder.name]
        assert result.returncode != 0 and f'Invariant {invariant} is violated.' in result.stdout, result.stdout
        print(folder.name, 'caught:', invariant, flush=True)
print('PASS: 10 model baselines and 19 targeted model mutations.')
