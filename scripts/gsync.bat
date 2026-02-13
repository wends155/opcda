@echo off
echo Syncing with remote...
git pull --rebase
if %ERRORLEVEL% NEQ 0 (
    echo Error pulling changes. Please resolve conflicts.
    exit /b %ERRORLEVEL%
)
git push
