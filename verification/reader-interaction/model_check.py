from pathlib import Path
import os
import subprocess
import tempfile

here = Path(__file__).resolve().parent
java = os.environ.get('JAVA', 'java')
jar = os.environ['TLA2TOOLS']
variants = [
    ('Reader', 'TypeOK', "page' = page + 1", "page' = 2"),
    ('Reader', 'SelectionValid', 'IF BadReset THEN selected ELSE -1', 'selected'),
    ('Reader', 'HistoryValid', 'Append(history,page)', 'Append(history,2)'),
    ('Styling', 'OrderedRanges', 'Start(k)>=last', 'TRUE'),
    ('Styling', 'SelectedOwner', 'Owner(k)=selected /\\ selected>=0', 'Owner(k)=selected'),
    ('Styling', 'OpenerSelection', '(~IsOpener(k) \\/ Owner(k)=selected)', 'TRUE'),
    ('Styling', 'LineIsolation', '~OtherLine(k)', 'TRUE'),
]
for module, invariant, before, after in variants:
    text = (here / f'{module}.tla').read_text()
    assert text.count(before) == 1, before
    config = 'SPECIFICATION Spec\n'
    if module == 'Reader':
        config += 'CONSTANT BadReset = FALSE\n'
    config += f'INVARIANT {invariant}\n'
    with tempfile.TemporaryDirectory(prefix='reader-model-mutant-') as tmp:
        scratch = Path(tmp)
        (scratch / f'{module}.tla').write_text(text.replace(before, after))
        (scratch / f'{module}.cfg').write_text(config)
        result = subprocess.run(
            [java, '-XX:+UseParallelGC', '-cp', jar, 'tlc2.TLC', '-workers', '1', module],
            cwd=scratch, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
        )
        assert f'Invariant {invariant} is violated' in result.stdout, result.stdout
    print(f'{module} {invariant}: detected intended violation', flush=True)
