#!/usr/bin/env sh
# Validates nginx/nginx.conf against the multi-node HA topology requirements
# defined in Task 12 of docs/superpowers/plans/2026-06-24-distributed-high-availability.md.
#
# Checks:
#   1. upstream drill_backend has at least two backend servers;
#   2. /api/ and /ws/ target the same upstream;
#   3. upstream servers target port 8080;
#   4. /api/ proxy_pass preserves /api/v1 (no trailing URI);
#   5. /ws/ proxy_read_timeout is 300s.
#
# Usage: scripts/check-nginx-config.sh [path/to/nginx.conf]
set -eu

CONF="${1:-nginx/nginx.conf}"
[ -f "$CONF" ] || { echo "FAIL: $CONF not found" >&2; exit 1; }

fail() { echo "FAIL: $1" >&2; exit 1; }

# Extract a location block by its exact argument using brace-depth matching.
extract_block() {
	awk -v want="location $1 " '
		$0 ~ want { inblock = 1; depth = 0 }
		inblock {
			print
			for (i = 1; i <= length($0); i++) {
				c = substr($0, i, 1)
				if (c == "{") depth++
				else if (c == "}") { depth--; if (depth == 0) { inblock = 0; next } }
			}
		}
	' "$CONF"
}

# 1. upstream drill_backend with >= 2 server directives.
upstream=$(awk '/upstream drill_backend \{/,/^\}/' "$CONF")
echo "$upstream" | grep -q '.' || fail "missing 'upstream drill_backend' block"
servers=$(echo "$upstream" | grep -cE '^[[:space:]]*server[[:space:]]+')
[ "$servers" -ge 2 ] || fail "expected >=2 upstream servers, found $servers"

# 3. upstream servers target port 8080.
port8080=$(echo "$upstream" | grep -cE ':8080\b')
[ "$port8080" -ge 2 ] || fail "upstream servers must target port 8080, found $port8080"

# 2 & 4. /api/ and /ws/ use the same upstream without a trailing URI.
api_block=$(extract_block '/api/')
ws_block=$(extract_block '/ws/')
api_pass=$(echo "$api_block" | grep -oE 'proxy_pass[[:space:]]+http://[^;]+;' | head -n1 | tr -s ' ')
ws_pass=$(echo "$ws_block" | grep -oE 'proxy_pass[[:space:]]+http://[^;]+;' | head -n1 | tr -s ' ')
[ -n "$api_pass" ] || fail "/api/ missing proxy_pass"
[ -n "$ws_pass" ] || fail "/ws/ missing proxy_pass"
[ "$api_pass" = "$ws_pass" ] || fail "/api/ and /ws/ target different upstreams: '$api_pass' vs '$ws_pass'"
[ "$api_pass" = "proxy_pass http://drill_backend;" ] || \
	fail "/api/ must be 'proxy_pass http://drill_backend;' (no trailing URI), got '$api_pass'"

# 5. WebSocket read timeout 300s.
echo "$ws_block" | grep -qE 'proxy_read_timeout[[:space:]]+300s' || \
	fail "/ws/ proxy_read_timeout must be 300s"

echo "OK: $CONF satisfies HA topology requirements"
