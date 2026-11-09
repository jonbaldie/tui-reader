from pathlib import Path
import os, shutil, subprocess, sys, tempfile
here=Path(__file__).resolve().parent
artifact='http:tla2tools[url=https://github.com/tlaplus/tlaplus/releases/download/v1.7.4/tla2tools.jar,checksum=sha256:936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88]@1.7.4'
jar=Path(subprocess.check_output(['mise','where',artifact],text=True).strip())/'tla2tools.jar'
java=subprocess.check_output(['mise','which','java','--tool','java@temurin-25'],text=True).strip()
env=dict(os.environ,JAVA=java,TLA2TOOLS=str(jar))
for model in sorted(here.glob('*/*.tla')):
 cfg=model.with_suffix('.cfg')
 if not cfg.exists():
  raise AssertionError(f'Missing configuration for {model}')
 with tempfile.TemporaryDirectory(prefix='reader-model-') as tmp:
  shutil.copy2(model,Path(tmp)/model.name)
  shutil.copy2(cfg,Path(tmp)/cfg.name)
  r=subprocess.run([java,'-XX:+UseParallelGC','-cp',str(jar),'tlc2.TLC','-workers','1','-config',cfg.name,model.name],cwd=tmp,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
  assert r.returncode==0,r.stdout
  print(model.relative_to(here),'PASS',flush=True)
for area,script,args in [('startup','verify.py',[]),('reader-interaction','model_check.py',[]),('reader-interaction','check.py',['--mutate']),('document-layout','model_check.py',[]),('document-layout','check.py',['--mutation']),('link-attribution','check.py',[])]:
 subprocess.run([sys.executable,str(here/area/script),*args],cwd=here.parent,env=env,check=True)
