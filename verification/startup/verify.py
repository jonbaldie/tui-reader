from pathlib import Path
import subprocess, sys
here=Path(__file__).resolve().parent
for script,args in [('model_check.py',[]),('check.py',[]),('check.py',['--mutation']),('check.py',[])]:
 subprocess.run([sys.executable,str(here/script),*args],check=True)
