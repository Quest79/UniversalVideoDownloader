$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$Bundle = Split-Path -Parent $MyInvocation.MyCommand.Path
$Root = Join-Path $env:LOCALAPPDATA 'UniversalVideoDownloader'
$HostExe = Join-Path $Root 'uvd-host.exe'
$YtDlp = Join-Path $Root 'yt-dlp.exe'
$Ffmpeg = Join-Path $Root 'ffmpeg.exe'
$Ffprobe = Join-Path $Root 'ffprobe.exe'
$Deno = Join-Path $Root 'deno.exe'
$HostName = 'universal_video_downloader_v44'
$Manifest = Join-Path $Root ($HostName + '.json')

Write-Host ''
Write-Host 'Universal Video Downloader v4.4 backend installer' -ForegroundColor Cyan
Write-Host 'This installs the local yt-dlp + FFmpeg helper. No admin rights are required.'
Write-Host ''

New-Item -ItemType Directory -Force -Path $Root | Out-Null
# Stop an older persistent helper so Windows can replace the executable.
Get-Process -Name 'uvd-host' -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Milliseconds 250
Copy-Item -Force (Join-Path $Bundle 'helper\uvd-host.exe') $HostExe

Write-Host '[1/6] Downloading/updating yt-dlp...'
Invoke-WebRequest -UseBasicParsing -Uri 'https://github.com/yt-dlp/yt-dlp-nightly-builds/releases/latest/download/yt-dlp.exe' -OutFile $YtDlp

if (-not (Test-Path $Ffmpeg) -or -not (Test-Path $Ffprobe)) {
    Write-Host '[2/6] Downloading FFmpeg essentials (one-time download, about 104 MB)...'
    $Zip = Join-Path $env:TEMP 'uvd-ffmpeg-release-essentials.zip'
    $Extract = Join-Path $env:TEMP 'uvd-ffmpeg-extract'
    Remove-Item -Recurse -Force $Extract -ErrorAction SilentlyContinue
    Remove-Item -Force $Zip -ErrorAction SilentlyContinue
    Invoke-WebRequest -UseBasicParsing -Uri 'https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip' -OutFile $Zip
    Expand-Archive -Path $Zip -DestinationPath $Extract -Force
    $FoundFfmpeg = Get-ChildItem -Path $Extract -Recurse -Filter 'ffmpeg.exe' | Select-Object -First 1
    $FoundFfprobe = Get-ChildItem -Path $Extract -Recurse -Filter 'ffprobe.exe' | Select-Object -First 1
    if (-not $FoundFfmpeg -or -not $FoundFfprobe) { throw 'FFmpeg archive did not contain ffmpeg.exe and ffprobe.exe.' }
    Copy-Item -Force $FoundFfmpeg.FullName $Ffmpeg
    Copy-Item -Force $FoundFfprobe.FullName $Ffprobe
    Remove-Item -Recurse -Force $Extract -ErrorAction SilentlyContinue
    Remove-Item -Force $Zip -ErrorAction SilentlyContinue
} else {
    Write-Host '[2/6] FFmpeg already installed.'
}

if (-not (Test-Path $Deno)) {
    Write-Host '[3/6] Downloading Deno for full YouTube support...'
    $DenoZip = Join-Path $env:TEMP 'uvd-deno.zip'
    $DenoExtract = Join-Path $env:TEMP 'uvd-deno-extract'
    Remove-Item -Recurse -Force $DenoExtract -ErrorAction SilentlyContinue
    Remove-Item -Force $DenoZip -ErrorAction SilentlyContinue
    Invoke-WebRequest -UseBasicParsing -Uri 'https://github.com/denoland/deno/releases/latest/download/deno-x86_64-pc-windows-msvc.zip' -OutFile $DenoZip
    Expand-Archive -Path $DenoZip -DestinationPath $DenoExtract -Force
    $FoundDeno = Get-ChildItem -Path $DenoExtract -Recurse -Filter 'deno.exe' | Select-Object -First 1
    if (-not $FoundDeno) { throw 'Deno archive did not contain deno.exe.' }
    Copy-Item -Force $FoundDeno.FullName $Deno
    Remove-Item -Recurse -Force $DenoExtract -ErrorAction SilentlyContinue
    Remove-Item -Force $DenoZip -ErrorAction SilentlyContinue
} else {
    Write-Host '[3/6] Deno already installed.'
}

Write-Host '[4/6] Registering Firefox native-messaging helper...'
# Remove stale registrations from earlier builds so Firefox cannot accidentally
# launch an incompatible old protocol host.
Remove-Item -Path 'HKCU:\Software\Mozilla\NativeMessagingHosts\universal_video_downloader' -Recurse -Force -ErrorAction SilentlyContinue
$NativeManifest = [ordered]@{
    name = $HostName
    description = 'Universal Video Downloader local yt-dlp helper'
    path = $HostExe
    type = 'stdio'
    allowed_extensions = @('universal-video-downloader@local')
}
$ManifestJson = $NativeManifest | ConvertTo-Json -Depth 4
[System.IO.File]::WriteAllText($Manifest, $ManifestJson, (New-Object System.Text.UTF8Encoding($false)))
$Reg = 'HKCU:\Software\Mozilla\NativeMessagingHosts\' + $HostName
New-Item -Path $Reg -Force | Out-Null
Set-Item -Path $Reg -Value $Manifest

Write-Host '[5/6] Verifying backend executable...'
$HostResult = & $HostExe --self-test
if ($LASTEXITCODE -ne 0) { throw 'Native helper self-test failed.' }
$YtVersion = & $YtDlp --version
$FfVersion = (& $Ffmpeg -version 2>$null | Select-Object -First 1)
$DenoVersion = (& $Deno --version 2>$null | Select-Object -First 1)


Write-Host '[6/6] Testing Firefox native-messaging protocol...'
$psi = New-Object System.Diagnostics.ProcessStartInfo
$psi.FileName = $HostExe
$psi.UseShellExecute = $false
$psi.RedirectStandardInput = $true
$psi.RedirectStandardOutput = $true
$psi.RedirectStandardError = $true
$psi.CreateNoWindow = $true
$proc = New-Object System.Diagnostics.Process
$proc.StartInfo = $psi
if (-not $proc.Start()) { throw 'Could not start native helper for protocol test.' }
$requestJson = '{"action":"ping"}'
$requestBytes = [System.Text.Encoding]::UTF8.GetBytes($requestJson)
$lenBytes = [System.BitConverter]::GetBytes([UInt32]$requestBytes.Length)
$proc.StandardInput.BaseStream.Write($lenBytes, 0, 4)
$proc.StandardInput.BaseStream.Write($requestBytes, 0, $requestBytes.Length)
$proc.StandardInput.BaseStream.Flush()
$proc.StandardInput.Close()
$header = New-Object byte[] 4
$read = $proc.StandardOutput.BaseStream.Read($header, 0, 4)
if ($read -ne 4) {
    $stderr = $proc.StandardError.ReadToEnd()
    throw "Native helper protocol test returned no header. $stderr"
}
$length = [System.BitConverter]::ToUInt32($header, 0)
if ($length -lt 2 -or $length -gt 1048576) { throw "Native helper returned invalid message length: $length" }
$payload = New-Object byte[] $length
$offset = 0
while ($offset -lt $length) {
    $n = $proc.StandardOutput.BaseStream.Read($payload, $offset, $length - $offset)
    if ($n -le 0) { break }
    $offset += $n
}
$proc.WaitForExit(5000) | Out-Null
if ($offset -ne $length) { throw 'Native helper protocol response was truncated.' }
$reply = [System.Text.Encoding]::UTF8.GetString($payload) | ConvertFrom-Json
if (-not $reply.ok -or $reply.backend_version -ne '4.4.0') {
    throw "Wrong native helper response/version: $($reply | ConvertTo-Json -Compress)"
}

Write-Host ''
Write-Host 'Backend installed successfully.' -ForegroundColor Green
Write-Host "yt-dlp: $YtVersion"
Write-Host "$FfVersion"
Write-Host "$DenoVersion"
Write-Host "Location: $Root"
Write-Host ''
Write-Host 'Now load extension\manifest.json in Firefox about:debugging -> This Firefox -> Load Temporary Add-on.' -ForegroundColor Yellow
