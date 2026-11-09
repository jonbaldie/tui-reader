from pathlib import Path
import argparse, shutil, subprocess, tempfile
p=argparse.ArgumentParser()
p.add_argument('--mutation',action='store_true')
a=p.parse_args()
root=Path(__file__).resolve().parents[2]
with tempfile.TemporaryDirectory(prefix='reader-startup-') as tmp:
    dst=Path(tmp)
    for name in ('main.go','go.mod','go.sum'):
        shutil.copy2(root/name,dst/name)
    shutil.copytree(root/'internal',dst/'internal')
    shutil.copy2(Path(__file__).with_name('args_test.go.txt'),dst/'formal_test.go')
    if a.mutation:
        source=dst/'main.go'
        text=source.read_text()
        assert text.count('parsed.dumpPages = pages')==1
        source.write_text(text.replace('parsed.dumpPages = pages','parsed.dumpPages = 0'))
    r=subprocess.run(['go','test','-count=1','-run','^TestFormalArguments$','.'],cwd=dst, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(r.stdout)
    if a.mutation:
        assert r.returncode!=0 and '--- FAIL: TestFormalArguments' in r.stdout and 'got' in r.stdout and 'want' in r.stdout, 'Expected parser assertion did not reject mutation: '+r.stdout
        print('Detected scratch Go mutation: counted dump ignored')
    else:
        assert r.returncode==0,'Actual Go argument fold failed'
