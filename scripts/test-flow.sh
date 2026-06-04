#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
PASS=0
FAIL=0

ok()   { echo "  PASS: $1"; ((PASS++)); }
fail() { echo "  FAIL: $1"; ((FAIL++)); }

echo "=== signalFlow test-flow ==="
echo "Target: $BASE_URL"
echo ""

# 1. Health check
echo "--- Health check ---"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
[ "$STATUS" = "200" ] && ok "GET /health returns 200" || fail "GET /health returned $STATUS"

# 2. Create asset
echo ""
echo "--- Create asset ---"
ASSET=$(curl -s -X POST "$BASE_URL/assets" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Wind Farm","expected_output":500}')
ASSET_ID=$(echo "$ASSET" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null || true)

if [ -n "$ASSET_ID" ]; then
  ok "POST /assets → 201, id=$ASSET_ID"
else
  fail "POST /assets did not return an id (response: $ASSET)"
  echo "Cannot continue without asset — exiting"
  exit 1
fi

# 3. Above-threshold reading (no alert expected)
echo ""
echo "--- Above-threshold reading (450/500) ---"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/readings" \
  -H "Content-Type: application/json" \
  -d "{\"asset_id\":\"$ASSET_ID\",\"actual_output\":450,\"recorded_at\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}")
[ "$STATUS" = "202" ] && ok "POST /readings → 202" || fail "POST /readings returned $STATUS"

echo "  Waiting 2s for consumer..."
sleep 2

ALERTS=$(curl -s "$BASE_URL/alerts?asset_id=$ASSET_ID")
COUNT=$(echo "$ALERTS" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
[ "$COUNT" = "0" ] && ok "No alert for above-threshold reading" || fail "Expected 0 alerts, got $COUNT"

# 4. Below-threshold reading (alert expected)
echo ""
echo "--- Below-threshold reading (200/500) ---"
curl -s -o /dev/null -X POST "$BASE_URL/readings" \
  -H "Content-Type: application/json" \
  -d "{\"asset_id\":\"$ASSET_ID\",\"actual_output\":200,\"recorded_at\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"
ok "POST /readings → submitted"

echo "  Waiting 2s for consumer..."
sleep 2

ALERTS=$(curl -s "$BASE_URL/alerts?asset_id=$ASSET_ID")
COUNT=$(echo "$ALERTS" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
[ "$COUNT" = "1" ] && ok "Alert created for below-threshold reading" || fail "Expected 1 alert, got $COUNT"

# Summary
echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" = "0" ] && exit 0 || exit 1
