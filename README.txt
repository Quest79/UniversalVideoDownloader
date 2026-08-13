Universal Video Downloader v4.4

IMPORTANT: Run INSTALL_HELPER.bat once before loading the extension.
The v4.4 extension intentionally uses a new native-host name so it cannot accidentally talk to an older incompatible helper.

UNIVERSAL VIDEO DOWNLOADER v4.4 - FIREFOX 142+ / WINDOWS
===================================================

ARCHITECTURE
Firefox extension -> one-shot Native Messaging helper -> yt-dlp/FFmpeg -> Windows Downloads

This avoids saving browser blob URLs, byte-range fragments, HLS/DASH manifests, or video-only tracks as if they were finished files.

INSTALL / UPGRADE
1. Double-click INSTALL_HELPER.bat.
   It installs/updates:
     - uvd-host.exe
     - current yt-dlp nightly
     - FFmpeg + ffprobe
     - Deno (required by current yt-dlp for full YouTube extraction)
   under:
     %LOCALAPPDATA%\UniversalVideoDownloader

2. Install the Mozilla-signed extension from addons.mozilla.org when its listing
   is available. For local development only, use Firefox -> about:debugging ->
   This Firefox -> Load Temporary Add-on, then select extension\manifest.json.

If upgrading from any older build, run INSTALL_HELPER.bat again because the helper changed.

USE
Right-click a video -> Download video.
The submenu is populated from formats reported by yt-dlp, such as:
  - Best quality + audio
  - Best compatible MP4
  - 2160p / 1440p / 1080p / 720p / etc.
  - Audio only

SAVE FOLDER
  C:\Users\<your Windows username>\Downloads

V4.4 FIXES
- New native-host name: v4.4 cannot accidentally connect to an older incompatible helper.
- Removed the persistent native connection that produced "Native helper disconnected" failures.
- Probe requests now use one-shot native messages.
- Downloads stay inside the helper until yt-dlp/FFmpeg actually finishes, so Firefox cannot kill a spawned child after an early response.
- The extension verifies backend_version=4.4.0 on every reply.
- INSTALL_HELPER.bat now tests the actual Firefox length-prefixed native-messaging protocol before reporting success.
- Keeps the v4.3 Deno, cookie, browser-impersonation, range-cleaning, and real-error improvements.

LOGS
Most recent download:
  %LOCALAPPDATA%\UniversalVideoDownloader\last-download.log
Most recent failed scan and its fallbacks:
  %LOCALAPPDATA%\UniversalVideoDownloader\last-probe.log

LIMITS
- DRM-protected streams are not bypassed.
- Login/private media is only available when your Firefox session itself has access and yt-dlp can use that session.
- Sites change. UPDATE_BACKEND.bat updates the extraction backend.
- Firefox 142 or later is required so Firefox can display the extension's built-in data-transmission disclosure on every supported Firefox platform without compatibility warnings. The add-on itself is distributed for Windows desktop only because it requires a Windows native helper.
- The legacy included XPI is an unsigned development artifact. The AMO upload package is generated under artifacts by build-amo-package.ps1 and must be signed by Mozilla before normal installation.
