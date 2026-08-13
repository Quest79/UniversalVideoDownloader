Universal Video Downloader v4.4
================================

A Firefox 142+ extension for Windows that downloads user-selected online video
through a local yt-dlp and FFmpeg companion.

FOR USERS
---------

1. Install the signed Firefox extension from addons.mozilla.org when its listing
   is available.
2. Download the Windows companion ZIP from:
   https://github.com/Quest79/UniversalVideoDownloader/releases/latest
3. Extract the ZIP and run INSTALL_HELPER.bat once.

The installer requires no administrator rights. It installs the native helper,
yt-dlp, FFmpeg, ffprobe, and Deno under:

  %LOCALAPPDATA%\UniversalVideoDownloader

To update yt-dlp and Deno later, run UPDATE_BACKEND.bat. To remove the companion
and its local logs, run UNINSTALL_HELPER.bat.

USING THE EXTENSION
-------------------

1. Right-click a video.
2. Open Download video with local helper.
3. Click Scan with local helper.
4. After the scan completes, right-click the video again and choose a quality.

Available choices can include best quality with audio, compatible MP4, individual
resolutions, and audio only. Finished files are saved in the Windows Downloads
folder.

HOW IT WORKS
------------

  Firefox extension
    -> one-shot Native Messaging request
    -> uvd-host.exe
    -> yt-dlp and FFmpeg
    -> Windows Downloads

The native helper is necessary for segmented HLS/DASH media, separate video and
audio tracks, format merging, and the site extractors maintained by yt-dlp.

REPOSITORY LAYOUT
-----------------

  extension\       Firefox extension source and icons
  helper-src\      Go source and tests for the native helper
  helper\          Prebuilt Windows amd64 native helper
  install-helper.ps1
                   Companion installer implementation
  INSTALL_HELPER.bat
                   User-facing installer launcher
  UPDATE_BACKEND.bat
                   Updates yt-dlp and Deno
  UNINSTALL_HELPER.bat
                   Removes the installed companion
  build-amo-package.ps1
                   Builds Firefox and GitHub upload artifacts

BUILDING
--------

Requirements for rebuilding the native helper:

  - Go 1.23 or later
  - Node.js 20 or later with web-ext available for Mozilla validation

Build and test the Windows helper from the repository root:

  cd helper-src
  go test ./...
  go build -trimpath -o ..\helper\uvd-host.exe .
  cd ..

Build the Firefox upload, Mozilla reviewer source, and Windows companion ZIPs:

  powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\build-amo-package.ps1

Generated files are written under artifacts\ and UPLOAD_READY\ and are excluded
from Git.

PRIVACY AND SECURITY BEHAVIOR
-----------------------------

- Scanning starts only after the user clicks the labeled scan command.
- Only HTTPS page and media URLs are sent to the local helper.
- Native-helper actions are disabled in Firefox private windows because the
  helper keeps local troubleshooting logs.
- The helper first tries public access. If necessary, yt-dlp may use applicable
  cookies from the user's normal Firefox session for media the user can access.
- No analytics, advertising, or developer-operated download server is used.

Local troubleshooting logs:

  %LOCALAPPDATA%\UniversalVideoDownloader\last-download.log
  %LOCALAPPDATA%\UniversalVideoDownloader\last-probe.log

LIMITS
------

- DRM is not bypassed.
- Login-gated media works only when the user already has access and yt-dlp can
  use that Firefox session.
- Site behavior changes over time; use UPDATE_BACKEND.bat to update yt-dlp.
- The companion currently supports Windows amd64 only.

LOCAL EXTENSION DEVELOPMENT
---------------------------

In Firefox, open about:debugging, choose This Firefox, select Load Temporary
Add-on, and open extension\manifest.json.
