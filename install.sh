#!/usr/bin/env sh
# Install kubot from GitHub Releases. Usage:
#   curl -fsSL https://raw.githubusercontent.com/kevinrst/kubot/main/install.sh | sh
#
# Env: KUBOT_VERSION (default: latest release tag), KUBOT_INSTALL_DIR
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

sha256check() {
  # sha256sum (coreutils/Linux, Git Bash) or shasum (macOS default).
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c -
  else
    shasum -a 256 -c -
  fi
}

resolve_tag() {
  if [ "$VERSION" != "latest" ]; then echo "$VERSION"; return; fi
  # Follow the /latest redirect to the newest tag without needing the API.
  curl -fsSL -o /dev/null -w "%{url_effective}\n" "https://github.com/$REPO/releases/latest" | sed 's#.*/tag/##'
}

main() {
  need() { command -v "$1" >/dev/null 2>&1 || { echo "need $1" >&2; exit 1; }; }
  need curl
  if [ "$(os)" = "windows" ]; then need unzip; else need tar; fi

  tag="$(resolve_tag)"
  # Release assets strip the leading v (kubot_0.1.0_..., not kubot_v0.1.0_...).
  ver="${tag#v}"
  plat="$(os)_$(arch)"
  if [ "$(os)" = "windows" ]; then
    arc="zip"; bin_name="kubot.exe"; dest_name="kubot.exe"
  else
    arc="tar.gz"; bin_name="kubot"; dest_name="kubot"
  fi
  base="kubot_${ver}_${plat}"
  url="https://github.com/$REPO/releases/download/${tag}/${base}.${arc}"
  tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT INT TERM

  echo "downloading kubot $tag ($plat)…"
  curl -fsSL -o "$tmp/${base}.${arc}" "$url"
  curl -fsSL -o "$tmp/checksums.txt" "https://github.com/$REPO/releases/download/${tag}/checksums.txt"

  (cd "$tmp" && grep "  ${base}.${arc}\$" checksums.txt | sha256check) || {
    echo "checksum mismatch — aborted" >&2; exit 1
  }

  if [ "$arc" = "zip" ]; then
    (cd "$tmp" && unzip -o -q "${base}.zip")
  else
    tar -xzf "$tmp/${base}.tar.gz" -C "$tmp"
  fi
  [ -f "$tmp/$bin_name" ] || { echo "archive missing $bin_name" >&2; exit 1; }

  mkdir -p "$INSTALL_DIR"
  install -m 0755 "$tmp/$bin_name" "$INSTALL_DIR/$dest_name"
  echo "installed to $INSTALL_DIR/$dest_name"
  "$INSTALL_DIR/$dest_name" --version
}

main
