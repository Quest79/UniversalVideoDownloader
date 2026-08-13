Universal Video Downloader - Windows Companion
==============================================

This is the Windows companion for Quest79's Universal Video Downloader Firefox
add-on. It is not the Firefox add-on itself.

The Firefox add-on provides the right-click video scan and download controls.
This companion lets the add-on use yt-dlp and FFmpeg on the computer to find
available video qualities, download streams, and merge separate video and audio.

FIREFOX ADD-ON
--------------

The Firefox add-on has not been published on addons.mozilla.org yet. Its Mozilla
listing link will be added here after Mozilla approves and publishes it.

Both the Firefox add-on and this Windows companion are required.

INSTALL THE WINDOWS COMPANION
-----------------------------

1. Download the latest companion ZIP:
   https://github.com/Quest79/UniversalVideoDownloader/releases/latest
2. Extract the ZIP.
3. Run INSTALL_HELPER.bat.

No administrator rights are required. The companion installs under:

  %LOCALAPPDATA%\UniversalVideoDownloader

Run UPDATE_BACKEND.bat to update yt-dlp and Deno.
Run UNINSTALL_HELPER.bat to remove the companion.

The companion supports Windows amd64. It does not bypass DRM.
