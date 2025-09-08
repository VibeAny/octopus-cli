# Chocolatey Install Script
$ErrorActionPreference = 'Stop'

$packageName = 'octopus-cli'
$version = '0.0.1'
$url64 = "https://github.com/VibeAny/octopus-cli/releases/download/v$version/octopus-v$version-windows-amd64-20250908.d4d58298.exe"
$url32 = "https://github.com/VibeAny/octopus-cli/releases/download/v$version/octopus-v$version-windows-386-20250908.d4d58298.exe"
$urlARM64 = "https://github.com/VibeAny/octopus-cli/releases/download/v$version/octopus-v$version-windows-arm64-20250908.d4d58298.exe"

$packageArgs = @{
    packageName    = $packageName
    fileType       = 'exe'
    url            = $url32
    url64bit       = $url64
    softwareName   = 'Octopus CLI*'
    checksum       = 'REPLACE_WITH_32BIT_CHECKSUM'
    checksum64     = 'REPLACE_WITH_64BIT_CHECKSUM'
    checksumType   = 'sha256'
    checksumType64 = 'sha256'
    silentArgs     = '/S'
    validExitCodes = @(0)
}

# Detect ARM64 and use appropriate URL
if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64' -or $env:PROCESSOR_ARCHITEW6432 -eq 'ARM64') {
    $packageArgs.url64bit = $urlARM64
    $packageArgs.checksum64 = 'REPLACE_WITH_ARM64_CHECKSUM'
}

Install-ChocolateyPackage @packageArgs

# Add to PATH if not already there
$binPath = Join-Path $env:ChocolateyInstall 'bin'
if ($env:PATH -notlike "*$binPath*") {
    $env:PATH = "$env:PATH;$binPath"
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, 'Machine')
}