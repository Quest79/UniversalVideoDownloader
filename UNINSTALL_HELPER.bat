@echo off
reg delete "HKCU\Software\Mozilla\NativeMessagingHosts\universal_video_downloader_v44" /f >nul 2>&1
reg delete "HKCU\Software\Mozilla\NativeMessagingHosts\universal_video_downloader" /f >nul 2>&1
rmdir /s /q "%LOCALAPPDATA%\UniversalVideoDownloader" >nul 2>&1
echo Universal Video Downloader backend removed.
pause
