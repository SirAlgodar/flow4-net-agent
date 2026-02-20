Param(
  [string]$InstallDir = "$env:ProgramFiles\Flow4Network\agent"
)

$ErrorActionPreference = "Stop"

$Target = "windows-amd64"
$Url = "https://example.com/flow4-net-agent-$Target.zip"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$zipPath = Join-Path $env:TEMP "flow4-net-agent.zip"
Invoke-WebRequest -Uri $Url -OutFile $zipPath
Expand-Archive -Path $zipPath -DestinationPath $InstallDir -Force

Write-Host "installed to $InstallDir"
