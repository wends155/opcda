@echo off
if "%~1"=="" (
    echo Usage: gcom "Commit message"
    exit /b 1
)

git add .
git commit -m "%~1"
