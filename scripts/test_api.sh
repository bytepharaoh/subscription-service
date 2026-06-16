#!/bin/bash

set -e

BASE_URL="http://localhost:8080"
API="$BASE_URL/api/v1/subscriptions"
PASS=0
FAIL=0

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() {

  echo -e "${GREEN}✓ $1${NC}"

  PASS=$((PASS + 1))

}
fail() {

  echo -e "${RED}✗ $1${NC}"

  FAIL=$((FAIL + 1))

}
section() { echo -e "\n${YELLOW}── $1 ──${NC}"; }

# ── Health ────────────────────────────────────────────────────────────────────
section "Health Check"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
[ "$STATUS" = "200" ] && pass "GET /health → 200" || fail "GET /health → $STATUS"

# ── Create ────────────────────────────────────────────────────────────────────
section "Create Subscriptions"

USER_ID="60601fee-2bf1-4721-ae6f-7636e79a0cba"
USER_ID_2="11111111-1111-1111-1111-111111111111"

RES=$(curl -s -X POST "$API" \
  -H "Content-Type: application/json" \
  -d "{\"service_name\":\"Yandex Plus\",\"price\":400,\"user_id\":\"$USER_ID\",\"start_date\":\"01-2025\"}")
ID1=$(echo "$RES" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
[ -n "$ID1" ] && pass "POST /subscriptions → created (id: $ID1)" || fail "POST /subscriptions → failed: $RES"

RES=$(curl -s -X POST "$API" \
  -H "Content-Type: application/json" \
  -d "{\"service_name\":\"Netflix\",\"price\":799,\"user_id\":\"$USER_ID\",\"start_date\":\"03-2025\",\"end_date\":\"09-2025\"}")
ID2=$(echo "$RES" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
[ -n "$ID2" ] && pass "POST /subscriptions → created with end_date (id: $ID2)" || fail "POST /subscriptions → failed: $RES"

RES=$(curl -s -X POST "$API" \
  -H "Content-Type: application/json" \
  -d "{\"service_name\":\"Spotify\",\"price\":299,\"user_id\":\"$USER_ID_2\",\"start_date\":\"06-2025\"}")
ID3=$(echo "$RES" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
[ -n "$ID3" ] && pass "POST /subscriptions → created for user 2 (id: $ID3)" || fail "POST /subscriptions → failed: $RES"

# ── Validation errors ─────────────────────────────────────────────────────────
section "Validation"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API" \
  -H "Content-Type: application/json" \
  -d '{"service_name":"","price":0,"user_id":"bad-uuid","start_date":""}')
[ "$STATUS" = "400" ] && pass "POST invalid body → 400" || fail "POST invalid body → $STATUS"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API" \
  -H "Content-Type: application/json" \
  -d "{\"service_name\":\"Test\",\"price\":100,\"user_id\":\"$USER_ID\",\"start_date\":\"99-9999\"}")
[ "$STATUS" = "400" ] && pass "POST invalid date format → 400" || fail "POST invalid date format → $STATUS"

# ── GetByID ───────────────────────────────────────────────────────────────────
section "Get By ID"

RES=$(curl -s "$API/$ID1")
SVC=$(echo "$RES" | grep -o '"service_name":"[^"]*"' | cut -d'"' -f4)
[ "$SVC" = "Yandex Plus" ] && pass "GET /subscriptions/:id → correct record" || fail "GET /subscriptions/:id → $RES"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API/00000000-0000-0000-0000-000000000000")
[ "$STATUS" = "404" ] && pass "GET /subscriptions/:id non-existent → 404" || fail "GET /subscriptions/:id non-existent → $STATUS"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API/not-a-uuid")
[ "$STATUS" = "400" ] && pass "GET /subscriptions/:id bad uuid → 400" || fail "GET /subscriptions/:id bad uuid → $STATUS"

# ── Update ────────────────────────────────────────────────────────────────────
section "Update"

RES=$(curl -s -X PUT "$API/$ID1" \
  -H "Content-Type: application/json" \
  -d "{\"service_name\":\"Yandex Plus Updated\",\"price\":599,\"start_date\":\"01-2025\"}")
SVC=$(echo "$RES" | grep -o '"service_name":"[^"]*"' | cut -d'"' -f4)
[ "$SVC" = "Yandex Plus Updated" ] && pass "PUT /subscriptions/:id → updated" || fail "PUT /subscriptions/:id → $RES"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$API/00000000-0000-0000-0000-000000000000" \
  -H "Content-Type: application/json" \
  -d '{"service_name":"X","price":100,"start_date":"01-2025"}')
[ "$STATUS" = "404" ] && pass "PUT /subscriptions/:id non-existent → 404" || fail "PUT /subscriptions/:id non-existent → $STATUS"

# ── List ──────────────────────────────────────────────────────────────────────
section "List"

RES=$(curl -s "$API")
COUNT=$(echo "$RES" | grep -o '"id"' | wc -l | tr -d ' ')
[ "$COUNT" = "3" ] && pass "GET /subscriptions → all 3 records" || fail "GET /subscriptions → expected 3, got $COUNT"

RES=$(curl -s "$API?user_id=$USER_ID")
COUNT=$(echo "$RES" | grep -o '"id"' | wc -l | tr -d ' ')
[ "$COUNT" = "2" ] && pass "GET /subscriptions?user_id → filtered 2 records" || fail "GET /subscriptions?user_id → expected 2, got $COUNT"

RES=$(curl -s "$API?service_name=Spotify")
COUNT=$(echo "$RES" | grep -o '"id"' | wc -l | tr -d ' ')
[ "$COUNT" = "1" ] && pass "GET /subscriptions?service_name → filtered 1 record" || fail "GET /subscriptions?service_name → expected 1, got $COUNT"

RES=$(curl -s "$API?page=1&page_size=2")
COUNT=$(echo "$RES" | grep -o '"id"' | wc -l | tr -d ' ')
[ "$COUNT" = "2" ] && pass "GET /subscriptions?page=1&page_size=2 → 2 records" || fail "GET /subscriptions pagination → expected 2, got $COUNT"

# ── Total Cost ────────────────────────────────────────────────────────────────
section "Total Cost"

RES=$(curl -s "$API/total-cost?period_start=01-2025&period_end=12-2025")
TOTAL=$(echo "$RES" | grep -o '"total":[0-9]*' | cut -d':' -f2)
[ "$TOTAL" = "1697" ] && pass "GET /total-cost → correct total (1498 RUB)" || fail "GET /total-cost → expected 1498, got $TOTAL: $RES"

RES=$(curl -s "$API/total-cost?period_start=01-2025&period_end=12-2025&user_id=$USER_ID")
TOTAL=$(echo "$RES" | grep -o '"total":[0-9]*' | cut -d':' -f2)
[ "$TOTAL" = "1398" ] && pass "GET /total-cost?user_id → 1199 RUB" || fail "GET /total-cost?user_id → expected 1199, got $TOTAL"

RES=$(curl -s "$API/total-cost?period_start=01-2025&period_end=12-2025&service_name=Spotify")
TOTAL=$(echo "$RES" | grep -o '"total":[0-9]*' | cut -d':' -f2)
[ "$TOTAL" = "299" ] && pass "GET /total-cost?service_name → 299 RUB" || fail "GET /total-cost?service_name → expected 299, got $TOTAL"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API/total-cost?period_start=12-2025&period_end=01-2025")
[ "$STATUS" = "400" ] && pass "GET /total-cost invalid period → 400" || fail "GET /total-cost invalid period → $STATUS"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API/total-cost?period_start=bad&period_end=01-2025")
[ "$STATUS" = "400" ] && pass "GET /total-cost bad date → 400" || fail "GET /total-cost bad date → $STATUS"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API/total-cost?period_end=01-2025")
[ "$STATUS" = "400" ] && pass "GET /total-cost missing period_start → 400" || fail "GET /total-cost missing period_start → $STATUS"

# ── Delete ────────────────────────────────────────────────────────────────────
section "Delete"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$API/$ID2")
[ "$STATUS" = "204" ] && pass "DELETE /subscriptions/:id → 204" || fail "DELETE /subscriptions/:id → $STATUS"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API/$ID2")
[ "$STATUS" = "404" ] && pass "GET deleted record → 404" || fail "GET deleted record → $STATUS"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$API/00000000-0000-0000-0000-000000000000")
[ "$STATUS" = "404" ] && pass "DELETE non-existent → 404" || fail "DELETE non-existent → $STATUS"
# ── Cleanup: delete all existing subscriptions ────────────────────────────────
section "Cleanup"

IDS=$(curl -s "$API" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
for ID in $IDS; do
  curl -s -X DELETE "$API/$ID" > /dev/null
done
echo -e "${YELLOW}cleaned up existing records${NC}"
# ── Summary ───────────────────────────────────────────────────────────────────
echo -e "\n────────────────────────────────"
echo -e "${GREEN}PASSED: $PASS${NC}"
[ "$FAIL" -gt 0 ] && echo -e "${RED}FAILED: $FAIL${NC}" || echo -e "${GREEN}FAILED: $FAIL${NC}"
echo "────────────────────────────────"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
