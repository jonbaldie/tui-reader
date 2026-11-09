from pathlib import Path
import subprocess, tempfile, json, sys, shutil
root=Path(__file__).resolve().parents[2]
cmd=['go','test','-count=1','./verification/reader-interaction','./internal/tui']
subprocess.run(cmd,cwd=root,check=True)
if '--mutate' in sys.argv:
    source=root/'internal/tui/model.go'
    original=source.read_text()
    old='m.currentPage++\n\t\t\tm.selectedLink = -1'
    assert original.count(old)==1
    with tempfile.TemporaryDirectory(prefix='reader-mutant-') as tmp:
        tmp=Path(tmp)
        modified=tmp/'model.go'
        modified.write_text(original.replace(old,'m.currentPage++'))
        overlay=tmp/'overlay.json'
        overlay.write_text(json.dumps({'Replace':{str(source):str(modified)}}))
        result=subprocess.run(['go','test','-overlay',str(overlay),'-count=1','-run','TestFocusClearsOnMove','./verification/reader-interaction'],cwd=root,capture_output=True,text=True)
        print(result.stdout+result.stderr)
        assert result.returncode!=0 and 'page movement retained link focus' in result.stdout, 'mutation was not detected by the expected assertion'
    subprocess.run(cmd,cwd=root,check=True)
    print('PASS: real Go focus-reset mutation detected; unchanged source passes.')
with tempfile.TemporaryDirectory(prefix='reader-spans-') as tmp:
    target=Path(tmp)
    for name in ('go.mod','go.sum'):
        shutil.copy2(root/name,target/name)
    shutil.copytree(root/'internal',target/'internal')
    shutil.copy2(Path(__file__).with_name('styling_test.go.txt'),target/'internal/tui/formal_span_test.go')
    stylecmd=['go','test','-count=1','-run','^TestFormalLabelSpans$','./internal/tui']
    subprocess.run(stylecmd,cwd=target,check=True)
    if '--mutate' in sys.argv:
        source=target/'internal/tui/model.go'
        original=source.read_text()
        before='span.LinkIndex == selectedIndex && selectedIndex >= 0'
        assert original.count(before)==1
        source.write_text(original.replace(before,'span.LinkIndex == selectedIndex'))
        result=subprocess.run(stylecmd,cwd=target,capture_output=True,text=True)
        print(result.stdout+result.stderr)
        assert result.returncode!=0 and 'TestFormalLabelSpans/continuation_remains_unselected' in result.stdout and 'span styling differs' in result.stdout, 'expected continuation assertion did not reject the source mutation'
        source.write_text(original)
        subprocess.run(stylecmd,cwd=target,check=True)
        print('PASS: continuation styling mutation detected; restored source passes.')
