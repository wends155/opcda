#!/bin/sh
if [ -z "$1" ]; then
    echo "Usage: gcom \"Commit message\""
    exit 1
fi

git add .
git commit -m "$1"
