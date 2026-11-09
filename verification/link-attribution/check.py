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

with tempfile.TemporaryDirectory(prefix='reader-link-check-') as tmp:
    work=Path(tmp)
    shutil.copy(HERE/'LinkAttribution.tla', work)
    cfg=(HERE/'LinkAttribution.cfg').read_text()
    (work/'LinkAttribution.cfg').write_text(cfg)
    tlc=[os.environ.get('JAVA','java'),'-XX:+UseParallelGC','-cp',str(jar),'tlc2.TLC','-workers','1','-config','LinkAttribution.cfg','LinkAttribution.tla']
    run(tlc,work,'model')
    for constant,invariant in [('MutateFence','NoCodeLinks'),('MutateQueue','ConsumedOnce')]:
        (work/'LinkAttribution.cfg').write_text(cfg.replace(constant+' = FALSE',constant+' = TRUE'))
        run(tlc,work,constant,'Invariant '+invariant+' is violated')
    (work/'LinkAttribution.cfg').write_text(cfg)
    model=(HERE/'LinkAttribution.tla').read_text()
    for label,before,after,invariant in [
        ('order','Append(attached, sources[next+1].raw)','Append(attached, 99)','AttachmentOrder'),
        ('cache',"cached' = TRUE","cached' = FALSE",'CacheBeforeAttach'),
        ('complete','phase = "Attach" /\\ next = Len(sources)','phase = "Attach"','Complete')]:
        assert before in model, label
        (work/'LinkAttribution.tla').write_text(model.replace(before,after,1))
        run(tlc,work,'model-mutant-'+label,'Invariant '+invariant+' is violated')
    (work/'LinkAttribution.tla').write_text(model)
    shutil.copy(HERE/'LinkScanners.tla',work)
    for mode in ('Anchor','Inline','Grouping','SourceCursor'):
        shutil.copy(HERE/(mode+'.cfg'),work/'Scanner.cfg')
        command=[os.environ.get('JAVA','java'),'-XX:+UseParallelGC','-cp',str(jar),'tlc2.TLC','-workers','1','-config','Scanner.cfg','LinkScanners.tla']
        run(command,work,'scanner-'+mode)
    scanner=(HERE/'LinkScanners.tla').read_text()
    for mode,before,after,invariant in [
        ('Anchor','THEN Append(Append(output,"-"),"a")','THEN Append(output,"-")','Normalized'),
        ('Inline','opener=run THEN 0 ELSE opener','opener=run THEN 0 ELSE opener','ExactRunScope'),
        ('Grouping',"groups'=Append(groups,-1)","groups'=Append(groups,99)",'GroupsValid'),
        ('SourceCursor',"THEN 2 ELSE IF p>=10 THEN 1 ELSE 0","THEN 0 ELSE IF p>=10 THEN 1 ELSE 0",'CursorOwnsMatch')]:
        text=scanner.replace(before,after,1)
        config=(HERE/(mode+'.cfg')).read_text()
        if mode=='Inline': config=config.replace('MutateRun = FALSE','MutateRun = TRUE')
        (work/'LinkScanners.tla').write_text(text)
        (work/'Scanner.cfg').write_text(config)
        run(command,work,'scanner-mutant-'+mode,'Invariant '+invariant+' is violated')
    (work/'LinkScanners.tla').write_text(scanner)
    shutil.copy(HERE/'InlinePairs.tla',work)
    paircfg=(HERE/'InlinePairs.cfg').read_text()
    (work/'InlinePairs.cfg').write_text(paircfg)
    paircmd=[os.environ.get('JAVA','java'),'-XX:+UseParallelGC','-cp',str(jar),'tlc2.TLC','-workers','1','-config','InlinePairs.cfg','InlinePairs.tla']
    run(paircmd,work,'inline-pairs')
    (work/'InlinePairs.cfg').write_text(paircfg.replace('MutatePair = FALSE','MutatePair = TRUE'))
    run(paircmd,work,'inline-pairs-mutant','Invariant ExactPairs is violated')
    (work/'InlinePairs.cfg').write_text(paircfg)
    pairmodel=(HERE/'InlinePairs.tla').read_text()
    for label,before,after,rule in [('order','<<cursor,close>>','<<close,cursor>>','OrderedPairs'),('recovery','ELSE Candidates','ELSE {}','EOFRecovery')]:
        assert before in pairmodel
        (work/'InlinePairs.tla').write_text(pairmodel.replace(before,after,1))
        run(paircmd,work,'inline-pairs-mutant-'+label,'Invariant '+rule+' is violated')
    go=work/'go'
    (go/'internal/book').mkdir(parents=True)
    for name in ('go.mod','go.sum'):
        shutil.copy(REPO/name,go/name)
    for src in (REPO/'internal/book').glob('*.go'):
        if not src.name.endswith('_test.go'): shutil.copy(src,go/'internal/book'/src.name)
    shutil.copy(HERE/'links_test.go.txt',go/'internal/book/formal_links_test.go')
    command=['go','test','./internal/book','-run','^TestFormal','-count=1']
    run(command,go,'go-baseline')
    src=go/'internal/book/anchors.go'
    original=src.read_text()
    mutations=[('queue','reflowed.next[key] = next + 1','reflowed.next[key] = next','TestFormalAttributionBound'),('fence','if inFence {','if false && inFence {','got'),('anchor','pendingHyphen = false\n\t\t\tcontinue','pendingHyphen = true\n\t\t\tcontinue','TestFormalAnchorBound')]
    for label,before,after,message in mutations:
        assert before in original, label+' mutation target moved'
        src.write_text(original.replace(before,after,1))
        run(command,go,'go-mutant-'+label,message)
        src.write_text(original)
    run(command,go,'go-restored')
print('Evidence:',out)
