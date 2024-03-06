import subprocess
import os
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
            name = filename[:-4]
            print("Running", name)
            try:
                output = execute(['go','run', '.', name],60)
            except subprocess.TimeoutExpired:
                print('TIMEOUT')


run_all('taint')
run_all('valueflow')