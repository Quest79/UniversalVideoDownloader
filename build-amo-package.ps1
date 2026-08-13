$ErrorActionPreference = 'Stop'

$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$ExtensionRoot = Join-Path $ProjectRoot 'extension'
$Artifacts = Join-Path $ProjectRoot 'artifacts'
$AddonArchive = Join-Path $Artifacts 'universal-video-downloader-4.4.0-amo.zip'
$SourceArchive = Join-Path $Artifacts 'universal-video-downloader-4.4.0-reviewer-source.zip'
$CompanionArchive = Join-Path $Artifacts 'universal-video-downloader-4.4.0-windows-companion.zip'
$DevelopmentXpi = Join-Path $ProjectRoot 'UniversalVideoDownloader-v4.4.xpi'
$UploadRoot = Join-Path $ProjectRoot 'UPLOAD_READY'
$FirefoxUpload = Join-Path $UploadRoot 'FIREFOX_ADDON_HUB'
$GitHubUpload = Join-Path $UploadRoot 'GITHUB_RELEASE'
$SourceStage = Join-Path $env:TEMP 'uvd-amo-reviewer-source'
$CompanionStage = Join-Path $env:TEMP 'uvd-windows-companion'

New-Item -ItemType Directory -Force -Path $Artifacts | Out-Null
New-Item -ItemType Directory -Force -Path $FirefoxUpload | Out-Null
New-Item -ItemType Directory -Force -Path $GitHubUpload | Out-Null
Remove-Item -Force $AddonArchive, $SourceArchive, $CompanionArchive -ErrorAction SilentlyContinue
Remove-Item -Force (Join-Path $FirefoxUpload '*') -ErrorAction SilentlyContinue
Remove-Item -Force (Join-Path $GitHubUpload '*') -ErrorAction SilentlyContinue
Remove-Item -Recurse -Force $SourceStage -ErrorAction SilentlyContinue
Remove-Item -Recurse -Force $CompanionStage -ErrorAction SilentlyContinue

# AMO requires manifest.json at the archive root, not inside an extension folder.
Compress-Archive -Path (Join-Path $ExtensionRoot '*') -DestinationPath $AddonArchive -CompressionLevel Optimal
Copy-Item -Force $AddonArchive $DevelopmentXpi

New-Item -ItemType Directory -Force -Path $SourceStage | Out-Null
Copy-Item -Recurse -Force $ExtensionRoot (Join-Path $SourceStage 'extension')
Copy-Item -Recurse -Force (Join-Path $ProjectRoot 'helper-src') (Join-Path $SourceStage 'helper-src')
Copy-Item -Recurse -Force (Join-Path $ProjectRoot 'helper') (Join-Path $SourceStage 'helper')
Copy-Item -Force (Join-Path $ProjectRoot 'install-helper.ps1') $SourceStage
Copy-Item -Force (Join-Path $ProjectRoot 'INSTALL_HELPER.bat') $SourceStage
Copy-Item -Force (Join-Path $ProjectRoot 'UNINSTALL_HELPER.bat') $SourceStage
Copy-Item -Force (Join-Path $ProjectRoot 'UPDATE_BACKEND.bat') $SourceStage
Copy-Item -Force (Join-Path $ProjectRoot 'README.txt') (Join-Path $SourceStage 'README.txt')
Compress-Archive -Path (Join-Path $SourceStage '*') -DestinationPath $SourceArchive -CompressionLevel Optimal
Remove-Item -Recurse -Force $SourceStage

New-Item -ItemType Directory -Force -Path $CompanionStage | Out-Null
Copy-Item -Recurse -Force (Join-Path $ProjectRoot 'helper') (Join-Path $CompanionStage 'helper')
Copy-Item -Force (Join-Path $ProjectRoot 'install-helper.ps1') $CompanionStage
Copy-Item -Force (Join-Path $ProjectRoot 'INSTALL_HELPER.bat') $CompanionStage
Copy-Item -Force (Join-Path $ProjectRoot 'UNINSTALL_HELPER.bat') $CompanionStage
Copy-Item -Force (Join-Path $ProjectRoot 'UPDATE_BACKEND.bat') $CompanionStage
Copy-Item -Force (Join-Path $ProjectRoot 'README.txt') $CompanionStage
Compress-Archive -Path (Join-Path $CompanionStage '*') -DestinationPath $CompanionArchive -CompressionLevel Optimal
Remove-Item -Recurse -Force $CompanionStage

Copy-Item -Force $AddonArchive (Join-Path $FirefoxUpload '1-UPLOAD-TO-FIREFOX.zip')
Copy-Item -Force $SourceArchive (Join-Path $FirefoxUpload '2-SOURCE-FOR-MOZILLA-REVIEW.zip')
Copy-Item -Force $CompanionArchive (Join-Path $GitHubUpload 'universal-video-downloader-4.4.0-windows-companion.zip')

Write-Host "Created: $AddonArchive"
Write-Host "Created: $SourceArchive"
Write-Host "Created: $CompanionArchive"
Write-Host "Updated: $DevelopmentXpi"
Write-Host "Upload folders: $UploadRoot"
