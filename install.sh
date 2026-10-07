#!/usr/bin/env sh
# Install kubot from GitHub Releases. Usage:
#   curl -fsSL https://raw.githubusercontent.com/kevinrst/kubot/main/install.sh | sh
#
# Env: KUBOT_VERSION (default: latest release), KUBOT_INSTALL_DIR
#      (default: /usr/local/bin). Verifies the sha256 checksum; cosign
#      signature verification is planned, not yet shipped.
set -eu

REPO="kevinrst/kubot"
VERSION="${KUBOT_VERSION:-latest}"
INSTALL_DIR="${KUBOT_INSTALL_DIR:-/usr/local/bin}"

os() {
  case "$(uname -s)" in
    Linux)  echo linux ;;
    Darwin) echo darwin ;;
    MINGW*|MSYS*|CYGWIN*) echo windows ;;
    *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
  esac
}

arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    arm64|aarch64) echo arm64 ;;
    *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
  esac
}

ext() { [ "$(os)" = "windows" ] && echo zip || echo tar.gz; }

resolve_version() {
  if [ "$VERSION" != "latest" ]; then echo "$VERSION"; return; fi
  # Follow the /latest redirect to the newest tag without needing the API.
  curl -fsSL -o /dev/null -w "%{url_effective}\n" "https://github.com/$REPO/releases/latest" | sed 's#.*/tag/##'
}

main() {
  need() { command -v "$1" >/dev/null 2>&1 || { echo "need $1" >&2; exit 1; }; }
  need curl; need tar
  [ "$(os)" = "windows" ] && need unzip || true

  ver="$(resolve_version)"
  plat="$(os)_$(arch)"
  base="kubot_${ver}_${plat}"
  url="https://github.com/$REPO/releases/download/${ver}/${base}.$(ext)"
  tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT INT TERM

  echo "downloading kubot $ver ($plat)…"
  curl -fsSL -o "$tmp/pkg.$(ext)" "$url"
  curl -fsSL -o "$tmp/checksums.txt" "https://github.com/$REPO/releases/download/${ver}/checksums.txt"

  (cd "$tmp" && grep "  ${base}.$(ext)\$" checksums.txt | sha256sum -c -) || {
    echo "checksum mismatch — aborted" >&2; exit 1
  }

  if [ "$(ext)" = "zip" ]; then
    (cd "$tmp" && unzip -o -q "pkg.zip")
  else
    tar -xzf "$tmp/pkg.tar.gz" -C "$tmp"
  fi

  bin="$(find "$tmp" -name 'kubot*' -type f | head -n 1)"
  install -m 0755 "$bin" "$INSTALL_DIR/kubot"
  echo "installed to $INSTALL_DIR/kubot"
  "$INSTALL_DIR/kubot" --version
}

main
