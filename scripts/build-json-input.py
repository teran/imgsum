#!/usr/bin/env python
#
# Usage:
#   find <dir> -type f | build-json-input.py
#
# The script will return to STDOUT JSON document in a format for
#   `imgsum -json-input`
# Designed for testing purposes.
#

import json
import sys

if __name__ == '__main__':
    files = []
    for line in sys.stdin:
        files.append(line.strip())

    print json.dumps({'files': files}, indent=4)
