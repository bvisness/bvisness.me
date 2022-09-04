#!/usr/bin/env python3

import subprocess

sass = 'sass'

try:
    subprocess.run([sass, '--version'], stdout=subprocess.DEVNULL)
except FileNotFoundError:
    # windows jank...
    sass = 'sass.cmd'

try:
    subprocess.run([sass, '--version'], stdout=subprocess.DEVNULL)
except FileNotFoundError:
    print("sass is not installed. Run the following to install it:")
    print("npm install -g sass")
    exit()

subprocess.run([sass, 'scss/all.scss', 'site/style.css'])
