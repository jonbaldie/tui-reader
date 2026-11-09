from pathlib import Path
import subprocess, tempfile, shutil
root=Path(__file__).resolve().parent
tool='http:tla2tools[url=https://github.com/tlaplus/tlaplus/releases/download/v1.7.4/tla2tools.jar,checksum=sha256:936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88]@1.7.4'
jar=Path(subprocess.check_output(['mise','where',tool],text=True).strip())/'tla2tools.jar'
java=subprocess.check_output(['mise','which','java','--tool','java@temurin-25'],text=True).strip()
base=(root/'Startup.tla').read_text()
variants={
 'baseline':base,
 'TypeOK':base.replace("count' = count + 1","count' = 3",1),
 'LoadFailureCannotSucceed':base.replace("exitCode' = IF loadFailed THEN 1 ELSE 0","exitCode' = 0"),
 'DumpCannotRead':base.replace('IF dump THEN IF loadFailed\' THEN "Load error" ELSE "Dump complete"','IF dump THEN "Reading"'),
 'UsageFails':base.replace('phase\' = "Usage error" /\\ exitCode\' = 1','phase\' = "Usage error" /\\ exitCode\' = 0'),
}
for name,text in variants.items():
 with tempfile.TemporaryDirectory(prefix='startup-tlc-') as tmp:
  d=Path(tmp); (d/'Startup.tla').write_text(text); shutil.copy2(root/'Startup.cfg',d/'Startup.cfg')
  r=subprocess.run([java,'-XX:+UseParallelGC','-cp',str(jar),'tlc2.TLC','-workers','1','Startup'],cwd=d,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
  if name=='baseline': assert r.returncode==0,r.stdout
  else: assert f'Invariant {name} is violated' in r.stdout,r.stdout
  print(name+': '+('PASS' if name=='baseline' else 'detected intended violation'))
