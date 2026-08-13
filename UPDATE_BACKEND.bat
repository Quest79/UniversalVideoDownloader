@echo off
setlocal
set "ROOT=%LOCALAPPDATA%\UniversalVideoDownloader"
if not exist "%ROOT%\yt-dlp.exe" (
  echo Backend is not installed. Run INSTALL_HELPER.bat first.
  pause
  exit /b 1
)

echo Updating yt-dlp nightly...
"%ROOT%\yt-dlp.exe" --update-to nightly
if errorlevel 1 goto :fail

if exist "%ROOT%\deno.exe" (
  echo.
  echo Updating Deno...
  "%ROOT%\deno.exe" upgrade
)

echo.
echo Backend updated.
pause
exit /b 0

:fail
echo.
echo Backend update failed.
pause
exit /b 1
