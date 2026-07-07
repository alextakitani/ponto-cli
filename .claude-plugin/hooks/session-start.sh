#!/usr/bin/env bash
# session-start.sh - Ponto plugin liveness check
#
# Lightweight: one subprocess call. Context priming happens on first
# use via the /ponto skill, not here.

set -euo pipefail

if ! command -v ponto &>/dev/null; then
  cat << 'EOF'
<hook-output>
Ponto plugin active — CLI not found on PATH.
Install: https://github.com/alextakitani/ponto-cli#installation
</hook-output>
EOF
  exit 0
fi

auth_json=$(ponto auth status --json 2>/dev/null || echo '{}')

if ! command -v jq &>/dev/null; then
  cat << 'EOF'
<hook-output>
Ponto plugin active.
</hook-output>
EOF
  exit 0
fi

is_auth=false
if parsed_auth=$(echo "$auth_json" | jq -er '.data.authenticated' 2>/dev/null); then
  is_auth="$parsed_auth"
fi

if [[ "$is_auth" == "true" ]]; then
  cat << 'EOF'
<hook-output>
Ponto plugin active.
</hook-output>
EOF
else
  cat << 'EOF'
<hook-output>
Ponto plugin active — not authenticated.
Run: ponto setup
</hook-output>
EOF
fi
