@echo off
setlocal
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0install-helper.ps1"
if errorlevel 1 (
  echo.
  echo V4.4 INSTALL FAILED. The error is shown above.
  pause
  exit /b 1
)
echo.
pause
