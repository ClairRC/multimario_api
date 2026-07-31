#!/bin/bash
set -euo pipefail

# Run every sync script
for item in ./sync/*.py; do
    if [ -f "$item" ]; then
        python3 "$item"
    fi
done

# Import each generated JSON file
for item in ./sync/syncjson/*; do
    if [ -f "$item" ]; then
        python3 importer.py "$item"
    fi
done