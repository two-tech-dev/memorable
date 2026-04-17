#!/usr/bin/env bash
# Memorable — Auto-install MCP for AI agents
# Usage: ./scripts/install.sh [--agent cursor|claude|copilot|windsurf|all] [--config /path/to/config.yaml]
#
# Detects installed agents and registers Memorable as an MCP server.
# Run with --agent all (default) to configure every detected agent.

set -euo pipefail

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
RESET='\033[0m'

MEMORABLE_BIN=""
CONFIG_FLAG=""
TARGET_AGENT="all"

# ── Helpers ──────────────────────────────────────────────
info()  { echo -e "${CYAN}[memorable]${RESET} $*"; }
ok()    { echo -e "${GREEN}[✓]${RESET} $*"; }
warn()  { echo -e "${YELLOW}[!]${RESET} $*"; }
fail()  { echo -e "${RED}[✗]${RESET} $*"; }

usage() {
  cat <<EOF
${BOLD}Memorable MCP Installer${RESET}

Usage: $0 [OPTIONS]

Options:
  --agent   AGENT   Target agent: cursor, claude, copilot, windsurf, all (default: all)
  --config  PATH    Path to memorable.yaml (optional, passed via -config flag)
  --help            Show this help

Examples:
  $0                          # Install for all detected agents
  $0 --agent cursor           # Install for Cursor only
  $0 --config ~/memorable.yaml --agent claude
EOF
  exit 0
}

# ── Parse args ───────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --agent)  TARGET_AGENT="$2"; shift 2 ;;
    --config) CONFIG_FLAG="$2"; shift 2 ;;
    --help)   usage ;;
    *) fail "Unknown option: $1"; usage ;;
  esac
done

# ── Locate memorable binary ──────────────────────────────
find_binary() {
  if command -v memorable &>/dev/null; then
    MEMORABLE_BIN="memorable"
  elif [[ -f "./bin/memorable" ]]; then
    MEMORABLE_BIN="$(cd "$(dirname ./bin/memorable)" && pwd)/memorable"
  elif [[ -f "$GOPATH/bin/memorable" ]] 2>/dev/null; then
    MEMORABLE_BIN="$GOPATH/bin/memorable"
  elif [[ -f "$HOME/go/bin/memorable" ]]; then
    MEMORABLE_BIN="$HOME/go/bin/memorable"
  else
    fail "memorable binary not found. Run 'make build' or 'go install github.com/two-tech-dev/memorable/cmd/memorable@latest' first."
    exit 1
  fi
  ok "Found memorable at: ${MEMORABLE_BIN}"
}

# ── Build MCP JSON args ─────────────────────────────────
build_args() {
  if [[ -n "$CONFIG_FLAG" ]]; then
    echo '"-config", "'"$CONFIG_FLAG"'"'
  else
    echo ""
  fi
}

# ── JSON block for MCP config ────────────────────────────
mcp_json_block() {
  local args
  args=$(build_args)
  if [[ -n "$args" ]]; then
    cat <<EOF
    "memorable": {
      "command": "${MEMORABLE_BIN}",
      "args": [${args}]
    }
EOF
  else
    cat <<EOF
    "memorable": {
      "command": "${MEMORABLE_BIN}",
      "args": []
    }
EOF
  fi
}

# ── Inject into JSON config file ─────────────────────────
# Uses python/python3 for safe JSON manipulation (avail on most systems).
inject_mcp_config() {
  local config_file="$1"
  local servers_key="$2"  # "mcpServers" or "servers"

  local args_json="[]"
  if [[ -n "$CONFIG_FLAG" ]]; then
    args_json='["-config", "'"$CONFIG_FLAG"'"]'
  fi

  if [[ -f "$config_file" ]]; then
    # File exists — merge
    local py_cmd
    py_cmd=$(command -v python3 || command -v python || echo "")
    if [[ -z "$py_cmd" ]]; then
      warn "python3/python not found. Writing config manually."
      write_fresh_config "$config_file" "$servers_key"
      return
    fi

    $py_cmd - "$config_file" "$servers_key" "$MEMORABLE_BIN" "$args_json" <<'PYEOF'
import json, sys
config_file, servers_key, bin_path, args_json = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
try:
    with open(config_file, 'r') as f:
        data = json.load(f)
except (json.JSONDecodeError, FileNotFoundError):
    data = {}

if servers_key not in data:
    data[servers_key] = {}

args = json.loads(args_json)
entry = {"command": bin_path, "args": args}
if servers_key == "servers":
    entry["type"] = "stdio"

data[servers_key]["memorable"] = entry

with open(config_file, 'w') as f:
    json.dump(data, f, indent=2)
    f.write('\n')
PYEOF
  else
    write_fresh_config "$config_file" "$servers_key"
  fi
}

write_fresh_config() {
  local config_file="$1"
  local servers_key="$2"

  mkdir -p "$(dirname "$config_file")"

  local args_json="[]"
  if [[ -n "$CONFIG_FLAG" ]]; then
    args_json='["-config", "'"$CONFIG_FLAG"'"]'
  fi

  if [[ "$servers_key" == "servers" ]]; then
    cat > "$config_file" <<EOF
{
  "${servers_key}": {
    "memorable": {
      "type": "stdio",
      "command": "${MEMORABLE_BIN}",
      "args": ${args_json}
    }
  }
}
EOF
  else
    cat > "$config_file" <<EOF
{
  "${servers_key}": {
    "memorable": {
      "command": "${MEMORABLE_BIN}",
      "args": ${args_json}
    }
  }
}
EOF
  fi
}

# ── Agent installers ─────────────────────────────────────

install_cursor() {
  local config_dir="$HOME/.cursor"
  local config_file="$config_dir/mcp.json"
  info "Configuring Cursor..."
  inject_mcp_config "$config_file" "mcpServers"
  ok "Cursor: ${config_file}"
}

install_claude() {
  local config_dir
  case "$(uname -s)" in
    Darwin) config_dir="$HOME/Library/Application Support/Claude" ;;
    *)      config_dir="$HOME/.config/claude" ;;
  esac
  local config_file="$config_dir/claude_desktop_config.json"
  info "Configuring Claude Desktop / Claude Code..."
  inject_mcp_config "$config_file" "mcpServers"
  ok "Claude: ${config_file}"
}

install_copilot() {
  # VS Code settings — workspace-level .vscode/mcp.json or user-level
  local config_file=".vscode/mcp.json"
  info "Configuring VS Code (GitHub Copilot)..."
  inject_mcp_config "$config_file" "servers"
  ok "VS Code Copilot: ${config_file}"
}

install_windsurf() {
  local config_dir="$HOME/.codeium/windsurf"
  local config_file="$config_dir/mcp_config.json"
  info "Configuring Windsurf..."
  inject_mcp_config "$config_file" "mcpServers"
  ok "Windsurf: ${config_file}"
}

# ── Main ─────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${BOLD}  🧠 Memorable MCP Installer${RESET}"
  echo ""

  find_binary

  case "$TARGET_AGENT" in
    cursor)   install_cursor ;;
    claude)   install_claude ;;
    copilot)  install_copilot ;;
    windsurf) install_windsurf ;;
    all)
      install_cursor
      install_claude
      install_copilot
      install_windsurf
      ;;
    *)
      fail "Unknown agent: $TARGET_AGENT"
      echo "Supported: cursor, claude, copilot, windsurf, all"
      exit 1
      ;;
  esac

  echo ""
  ok "Done! Restart your agent to activate Memorable."
  echo ""
}

main
