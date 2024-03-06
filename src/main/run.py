import subprocess
import os, shutil
import os.path
from typing import Union

def execute(cmd: list[str], t: Union[None, int] = None) -> str:
    return subprocess.run(
        cmd,
        text = True,
        stdout = subprocess.PIPE,
        stderr = subprocess.STDOUT,
        timeout = t
    ).stdout


def run_all(path: str) -> None:
    for dirpath, dirnames, filenames in os.walk(path):
        for filename in filenames:
            print("Running", filename)
            try:
                output = execute(['go','run', '.', dirpath + '/' + filename],60)
            except subprocess.TimeoutExpired:
                print('TIMEOUT')

if os.path.isdir('taint-out'):
    shutil.rmtree('taint-out')
if os.path.isdir('valueflow-out'):
    shutil.rmtree('valueflow-out')
os.mkdir('taint-out') 
os.mkdir('valueflow-out') 
run_all('taint')
run_all('valueflow')