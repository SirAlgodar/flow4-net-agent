Param(
  [string]$InstallDir = "$env:ProgramFiles\Flow4Network\agent",
  [string]$OfflineDir = ""
)

# Instalador autossuficiente do flow4-net-agent para Windows
# - Usa binários estáticos pré-compilados
# - Suporta modo online (download) e offline (diretório local)
# - Faz validação pós-instalação com `flow4-net-agent ping-agent`

$ErrorActionPreference = "Stop"

function Write-Info($msg)  { Write-Host "[INFO] $msg"  -ForegroundColor Cyan }
function Write-Err($msg)   { Write-Host "[ERRO] $msg"  -ForegroundColor Red }
function Write-Warn($msg)  { Write-Host "[AVISO] $msg" -ForegroundColor Yellow }

if (-not (Test-Path $InstallDir)) {
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

$arch = $env:PROCESSOR_ARCHITECTURE
switch ($arch) {
  "AMD64" { $archTag = "amd64" }
  "ARM64" { $archTag = "arm64" }
  Default {
    Write-Err "Arquitetura não suportada: $arch"
    exit 1
  }
}

$binaryName    = "flow4-net-agent-windows-$archTag.exe"
$targetExePath = Join-Path $InstallDir "flow4-net-agent.exe"

Write-Info "Arquitetura detectada: $arch ($archTag)"
Write-Info "Diretório de instalação: $InstallDir"

if ($OfflineDir -and $OfflineDir.Trim() -ne "") {
  Write-Info "Modo offline habilitado. Usando pacotes em: $OfflineDir"
  $binPath = Join-Path $OfflineDir $binaryName
  $zipPath = Join-Path $OfflineDir ($binaryName + ".zip")

  if (Test-Path $binPath) {
    Write-Info "Encontrado binário estático: $binaryName"
    Copy-Item $binPath $targetExePath -Force
  } elseif (Test-Path $zipPath) {
    Write-Info "Encontrado pacote ZIP: $($zipPath)"
    $tmpDir = New-Item -ItemType Directory -Path ([System.IO.Path]::Combine($env:TEMP, [System.Guid]::NewGuid().ToString()))
    Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force
    $extracted = Join-Path $tmpDir $binaryName
    if (-not (Test-Path $extracted)) {
      Write-Err "Arquivo $binaryName não encontrado dentro do ZIP."
      exit 1
    }
    Copy-Item $extracted $targetExePath -Force
  } else {
    Write-Err "Não encontrei $binaryName nem $binaryName.zip em $OfflineDir"
    Write-Host "       Certifique-se de copiar o pacote correto para o diretório offline." -ForegroundColor Red
    exit 1
  }
} else {
  Write-Info "Modo online. Baixando binário do agente..."

  if (-not (Get-Command Invoke-WebRequest -ErrorAction SilentlyContinue)) {
    Write-Err "Invoke-WebRequest não está disponível. Use modo offline (-OfflineDir) ou atualize o PowerShell."
    exit 1
  }

  $url = "https://github.com/SirAlgodar/flow4-net-agent/releases/latest/download/$binaryName"
  Write-Info "Baixando: $url"

  $tmpFile = [System.IO.Path]::GetTempFileName()
  try {
    Invoke-WebRequest -Uri $url -OutFile $tmpFile -UseBasicParsing
  } catch {
    Write-Err "Falha ao baixar o binário do agente. $_"
    Write-Host "       Verifique a conexão de rede ou use modo offline (-OfflineDir)." -ForegroundColor Red
    exit 1
  }

  Copy-Item $tmpFile $targetExePath -Force
}

Write-Info "Binário instalado em: $targetExePath"

try {
  & $targetExePath ping-agent | Out-Null
} catch {
  Write-Err "O agente foi instalado, mas o comando 'ping-agent' falhou. $_"
  Write-Host "       Possíveis causas:" -ForegroundColor Red
  Write-Host "         - Binário corrompido (reinstale ou recopie o pacote)." -ForegroundColor Red
  Write-Host "         - Permissões de execução insuficientes." -ForegroundColor Red
  Write-Host "       Caminho do binário: $targetExePath" -ForegroundColor Red
  exit 1
}

Write-Host "[OK] flow4-net-agent instalado e validado com sucesso." -ForegroundColor Green
Write-Info  "Exemplo de uso: `"$targetExePath -json`""
