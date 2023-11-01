#!/usr/bin/env bash

set -euo pipefail

OUTPUT=$(golines --list-files $@)

if [ -n "$OUTPUT" ]; then
    echo "golines needs to be run on the following files:"
    echo "$OUTPUT"
    exit 1
fi

exit 0
