#!/usr/bin/env python3
import hashlib
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import urllib.request

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[1]
JAR_HASH = '936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88'
jar = Path(os.environ.get('TLA2TOOLS', Path.home()/'.cache/tui-reader-formal/tla2tools-1.7.4.jar'))
if not jar.exists():
    jar.parent.mkdir(parents=True, exist_ok=True)
    urllib.request.urlretrieve('https://github.com/tlaplus/tlaplus/releases/download/v1.7.4/tla2tools.jar', jar)
assert hashlib.sha256(jar.read_bytes()).hexdigest() == JAR_HASH, 'TLC checksum mismatch'
out = Path(os.environ.get('VERIFY_OUT', tempfile.mkdtemp(prefix='reader-links-evidence-'))).resolve()
out.mkdir(parents=True, exist_ok=True)

def run(command, cwd, name, expected=0):
    result = subprocess.run(command, cwd=cwd, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    (out/(name+'.txt')).write_text(result.stdout)
    if expected == 0:
        assert result.returncode == 0, result.stdout
    else:
        assert result.returncode != 0 and expected in result.stdout, result.stdout
    print(name, 'PASS' if expected == 0 else 'EXPECTED FAILURE')

def tlc(work, module, config):
    return [os.environ.get('JAVA','java'),'-XX:+UseParallelGC','-cp',str(jar),'tlc2.TLC','-workers','1','-config',config,module+'.tla']

with tempfile.TemporaryDirectory(prefix='reader-link-check-') as tmp:
    work=Path(tmp)
    models=['LinkAttribution','LinkScanners','InlinePairs','CodeCursor']
    for model in models:
        shutil.copy(HERE/(model+'.tla'),work)
        shutil.copy(HERE/(model+'.cfg'),work)
        run(tlc(work,model,model+'.cfg'),work,'model-'+model)
    mutations=[
        ('LinkAttribution','order','id |-> next,page |-> Owner(next)','id |-> 99,page |-> Owner(next)','OccurrenceOrder'),
        ('LinkAttribution','ownership','THEN Local(i) ELSE -1','THEN Local(i) ELSE Local(i)','SpanOwnership'),
        ('LinkAttribution','fallback','spans\'=spans \\cup Expected(next)','spans\'=spans \\cup {s \\in Expected(next): ~s.opener}','OwnerFallback'),
        ('LinkAttribution','complete','spans\'=spans \\cup Expected(next)','spans\'=spans','CompleteSpans'),
        ('LinkScanners','separator','THEN Append(Append(output,"-"),"a")','THEN Append(output,"-")','Normalized'),
        ('InlinePairs','pair','IF MutatePair THEN','IF TRUE THEN','ExactPairs'),
        ('InlinePairs','order','<<cursor,close>>','<<close,cursor>>','OrderedPairs'),
        ('InlinePairs','recovery','ELSE Candidates','ELSE {}','EOFRecovery'),
        ('CodeCursor','overlap','Spans[f][1]<a \\/ Spans[l-1][2]>b','Spans[f][1]<a /\\ Spans[l-1][2]>b','OverlapEquivalent'),
    ]
    for module,label,before,after,invariant in mutations:
        original=(HERE/(module+'.tla')).read_text()
        assert before in original, label+' model target moved'
        (work/(module+'.tla')).write_text(original.replace(before,after,1))
        cfg=(HERE/(module+'.cfg')).read_text()
        cfg=cfg[:cfg.index('INVARIANT')]+'INVARIANT '+invariant+'\n'
        (work/'Mutant.cfg').write_text(cfg)
        run(tlc(work,module,'Mutant.cfg'),work,'model-mutant-'+module+'-'+label,'Invariant '+invariant+' is violated')
        (work/(module+'.tla')).write_text(original)
    go=work/'go'
    (go/'internal/book').mkdir(parents=True)
    for name in ('go.mod','go.sum'):
        shutil.copy(REPO/name,go/name)
    for src in (REPO/'internal/book').glob('*.go'):
        if not src.name.endswith('_test.go'): shutil.copy(src,go/'internal/book'/src.name)
    shutil.copy(HERE/'links_test.go.txt',go/'internal/book/formal_links_test.go')
    command=['go','test','./internal/book','-run','^TestFormal','-count=1']
    run(command,go,'go-baseline')
    mutations=[
        ('book.go','occurrence','pages[pageIndex].Links = append(pages[pageIndex].Links, link)','pages[pageIndex].Links = append(pages[pageIndex].Links, link, link)','TestFormalPlacementBound'),
        ('anchors.go','continuation','index := -1','index := linkIndex','TestFormalPlacementBound'),
        ('anchors.go','fallback','if !visibleOnOwner {','if visibleOnOwner {','TestFormalPlacementBound'),
        ('anchors.go','cursor','return c.spans[c.first][0] < start || c.spans[c.last-1][1] > end','return c.spans[c.first][0] < start && c.spans[c.last-1][1] > end','TestFormalCodeCursor'),
        ('book.go','scanner','open = i','open = -1','TestFormalWrapScanner'),
    ]
    for filename,label,before,after,message in mutations:
        src=go/'internal/book'/filename
        original=src.read_text()
        assert before in original, label+' source mutation target moved'
        try:
            src.write_text(original.replace(before,after,1))
            run(command,go,'go-mutant-'+label,message)
        finally:
            src.write_text(original)
    run(command,go,'go-restored')
print('Evidence:',out)
