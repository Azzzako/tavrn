#!/usr/bin/env bash
# Usage: bash deploy/verify.sh [hostname]
# Checks that a tavrn deployment is healthy.
set -euo pipefail

HOST="${1:-tavrn.sh}"
OK=0
FAIL=0

pass() { echo "  [ok]  $*"; ((OK++)) || true; }
fail() { echo "  [!!]  $*"; ((FAIL++)) || true; }

echo "=== tavrn deployment check: $HOST ==="
echo

# 1. HTTPS + vanity import
echo "--- Go vanity import ---"
if curl -fsSL "https://${HOST}/?go-get=1" 2>/dev/null | grep -q "go-import"; then
  pass "https://${HOST}/?go-get=1 returns go-import meta tag"
else
  fail "https://${HOST}/?go-get=1 did not return go-import meta tag"
fi

# 2. TLS certificate
echo "--- TLS certificate ---"
EXPIRY=$(echo | openssl s_client -servername "$HOST" -connect "${HOST}:443" 2>/dev/null \
  | openssl x509 -noout -enddate 2>/dev/null | cut -d= -f2)
if [ -n "$EXPIRY" ]; then
  pass "TLS cert valid through: $EXPIRY"
else
  fail "Could not retrieve TLS certificate"
fi

# 3. SSH port 22 responds (tavrn server)
echo "--- SSH port 22 (tavrn server) ---"
BANNER=$(timeout 5 bash -c "echo '' | nc -w3 ${HOST} 22 2>/dev/null" || true)
if echo "$BANNER" | grep -qi "SSH"; then
  pass "Port 22 returns SSH banner"
else
  fail "Port 22 did not return SSH banner (tavrn server may be down)"
fi

# 4. Go module resolution (server binary)
echo "--- Module resolution ---"
if GONOSUMCHECK="*" GOFLAGS="-mod=mod" \
   go list -m -json -mod=mod "tavrn.sh/cmd/tavrn-admin@latest" 2>/dev/null | grep -q '"Path"'; then
  pass "go module tavrn.sh/cmd/tavrn-admin resolves"
else
  # Fallback: just check HTTPS returns the import page
  if curl -fsSL "https://${HOST}/cmd/tavrn-admin?go-get=1" 2>/dev/null | grep -q "go-import"; then
    pass "go vanity path /cmd/tavrn-admin resolves via HTTPS"
  else
    fail "tavrn.sh/cmd/tavrn-admin module path did not resolve"
  fi
fi

echo
echo "=== Results: ${OK} passed, ${FAIL} failed ==="
[ "$FAIL" -eq 0 ]
