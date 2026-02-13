#!/bin/sh
echo "Syncing with remote..."
git pull --rebase
if [ $? -ne 0 ]; then
    echo "Error pulling changes. Please resolve conflicts."
    exit 1
fi
git push
