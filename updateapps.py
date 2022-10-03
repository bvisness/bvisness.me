#!/usr/bin/env python3

import os
import subprocess

class dir:
    def __init__(self, path):
        self.path = path
        self.initial_dir = os.getcwd()
      
    def __enter__(self):
        os.chdir(self.path)
  
    def __exit__(self, type, value, traceback):
        os.chdir(self.initial_dir)

print('Updating git stuff...')
subprocess.run(['git', 'submodule', 'init'])
subprocess.run(['git', 'submodule', 'update', '--init', '--recursive'])
subprocess.run(['git', 'submodule', 'update', '--remote', '--recursive'])
subprocess.run(['git', 'submodule', 'foreach', '--recursive', 'git', 'clean', '-xfd'])
subprocess.run(['git', 'submodule', 'foreach', '--recursive', 'git', 'reset', '--hard'])
subprocess.run(['git', 'submodule', 'update', '--init', '--recursive'])
# yes, this is the most paranoid series of commands of all time.

print('Updating boggler...')
with dir('ext/boggler'):
    subprocess.run(['python3', 'buildwords.py'])
    subprocess.run(['python3', 'dist.py'])

print('Updating netsim...')
with dir('ext/netsim'):
    subprocess.run(['python3', 'build.py'])
