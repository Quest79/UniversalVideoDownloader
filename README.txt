Universal Video Downloader - Windows Companion
==============================================

This repository contains the Windows companion required by the Universal Video
Downloader Firefox add-on. The Firefox extension itself is distributed through
addons.mozilla.org.

INSTALL
-------

1. Download the companion ZIP from:
   https://github.com/Quest79/UniversalVideoDownloader/releases/latest
2. Extract the ZIP.
3. Run INSTALL_HELPER.bat.

No administrator rights are required. The companion installs under:

  %LOCALAPPDATA%\UniversalVideoDownloader

Run UPDATE_BACKEND.bat to update yt-dlp and Deno.
Run UNINSTALL_HELPER.bat to remove the companion.

The companion supports Windows amd64. It does not bypass DRM.
