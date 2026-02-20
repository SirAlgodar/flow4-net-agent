#!/usr/bin/env bash
set -euo pipefail

# Instalador autossuficiente do flow4-net-agent (Linux/macOS)
# - Usa binários estáticos pré-compilados
# - Suporta modo online (download) e offline (diretório local)
# - Faz validação pós-instalação com `flow4-net-agent ping-agent`

OFFLINE_DIR=""
INSTALL_DIR="$HOME/.flow4network/agent"
USE_DOCKER="0" # reservado para uso futuro (Docker/Podman)

print_usage() {
  cat <<EOF
Uso: install.sh [opções]

Opções:
  --offline-dir <caminho>   Diretório local contendo o binário ou pacote do agente
  --install-dir <caminho>   Diretório de instalação (default: $HOME/.flow4network/agent)
  -h, --help                Mostrar esta ajuda

Exemplos:
  # Instala usando download (online)
  ./install.sh

  # Instala usando pacote offline
  ./install.sh --offline-dir /media/pendrive/flow4-net-agent
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --offline-dir)
      OFFLINE_DIR="${2:-}"
      shift 2
      ;;
    --install-dir)
      INSTALL_DIR="${2:-}"
      shift 2
      ;;
    -h|--help)
      print_usage
      exit 0
      ;;
    *)
      echo "[ERRO] Opção desconhecida: $1" >&2
      print_usage
      exit 1
      ;;
  esac
done

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux*)  os_tag="linux" ;;
  darwin*) os_tag="darwin" ;;
  *)
    echo "[ERRO] Sistema operacional não suportado: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64) arch_tag="amd64" ;;
  arm64|aarch64) arch_tag="arm64" ;;
  *)
    echo "[ERRO] Arquitetura não suportada: $ARCH" >&2
    exit 1
    ;;
esac

BINARY_BASENAME="flow4-net-agent-${os_tag}-${arch_tag}"
TARGET_BIN="$INSTALL_DIR/flow4-net-agent"

ensure_prereqs_online() {
  if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
    echo "[ERRO] Nem curl nem wget encontrados. Instale um deles ou use modo offline (--offline-dir)." >&2
    exit 1
  fi
}

mkdir -p "$INSTALL_DIR"

echo "[INFO] SO detectado: $OS ($ARCH)"
echo "[INFO] Instalando em: $INSTALL_DIR"

if [[ -n "$OFFLINE_DIR" ]]; then
  echo "[INFO] Modo offline habilitado. Usando pacotes em: $OFFLINE_DIR"
  if [[ -f "$OFFLINE_DIR/$BINARY_BASENAME" ]]; then
    echo "[INFO] Encontrado binário estático: $BINARY_BASENAME"
    cp "$OFFLINE_DIR/$BINARY_BASENAME" "$TARGET_BIN"
  elif [[ -f "$OFFLINE_DIR/$BINARY_BASENAME.tar.gz" ]]; then
    echo "[INFO] Encontrado pacote: $BINARY_BASENAME.tar.gz"
    if ! command -v tar >/dev/null 2>&1; then
      echo "[ERRO] 'tar' não encontrado. Necessário para extrair pacote offline." >&2
      exit 1
    fi
    tar -xzf "$OFFLINE_DIR/$BINARY_BASENAME.tar.gz" -C "$INSTALL_DIR"
    mv "$INSTALL_DIR/$BINARY_BASENAME" "$TARGET_BIN"
  else
    echo "[ERRO] Não encontrei $BINARY_BASENAME nem $BINARY_BASENAME.tar.gz em $OFFLINE_DIR" >&2
    echo "       Certifique-se de copiar o pacote correto para o diretório offline." >&2
    exit 1
  fi
else
  echo "[INFO] Modo online. Baixando binário do agente..."
  ensure_prereqs_online

  URL="https://github.com/SirAlgodar/flow4-net-agent/releases/latest/download/${BINARY_BASENAME}"
  TMPFILE="$(mktemp)"

  echo "[INFO] Baixando: $URL"
  if command -v curl >/dev/null 2>&1; then
    if ! curl -fsSL "$URL" -o "$TMPFILE"; then
      echo "[ERRO] Falha ao baixar o binário do agente." >&2
      echo "       Verifique a conexão de rede ou use modo offline (--offline-dir)." >&2
      rm -f "$TMPFILE"
      exit 1
    fi
  else
    if ! wget -q "$URL" -O "$TMPFILE"; then
      echo "[ERRO] Falha ao baixar o binário do agente." >&2
      echo "       Verifique a conexão de rede ou use modo offline (--offline-dir)." >&2
      rm -f "$TMPFILE"
      exit 1
    fi
  fi

  mv "$TMPFILE" "$TARGET_BIN"
fi

chmod +x "$TARGET_BIN"

echo "[INFO] Binário instalado em: $TARGET_BIN"

# Tenta criar symlink em /usr/local/bin, se possível
if [[ -d "/usr/local/bin" && -w "/usr/local/bin" ]]; then
  ln -sf "$TARGET_BIN" /usr/local/bin/flow4-net-agent
  echo "[INFO] Link criado: /usr/local/bin/flow4-net-agent"
else
  echo "[AVISO] Não foi possível criar link em /usr/local/bin (sem permissão)."
  echo "        Adicione o diretório ao PATH manualmente, por exemplo:"
  echo "        export PATH=\"$INSTALL_DIR:\$PATH\"" 
fi

# Validação pós-instalação
echo "[INFO] Validando instalação (flow4-net-agent ping-agent)..."
if ! "$TARGET_BIN" ping-agent >/dev/null 2>&1; then
  echo "[ERRO] O agente foi instalado, mas o comando 'ping-agent' falhou." >&2
  echo "       Possíveis causas:" >&2
  echo "         - Binário corrompido (reinstale ou recopie o pacote)." >&2
  echo "         - Permissões de execução insuficientes." >&2
  echo "       Caminho do binário: $TARGET_BIN" >&2
  exit 1
fi

echo "[OK] flow4-net-agent instalado e validado com sucesso."
echo "[INFO] Exemplo de uso: flow4-net-agent -json"
