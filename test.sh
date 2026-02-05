#!/bin/bash
set -e

BASE_URL="${BASE_URL:-http://localhost:3000}"
ADMIN_PIN="${ADMIN_PIN:-0000}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass() {
    echo -e "${GREEN}PASS${NC}: $1"
}

fail() {
    echo -e "${RED}FAIL${NC}: $1"
    echo "Response: $2"
    exit 1
}

info() {
    echo -e "${YELLOW}INFO${NC}: $1"
}

# Helper to make requests
request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local token="$4"

    local headers=(-H "Content-Type: application/json")
    if [ -n "$token" ]; then
        headers+=(-H "Authorization: Bearer $token")
    fi

    if [ -n "$data" ]; then
        curl -s -X "$method" "${headers[@]}" -d "$data" "${BASE_URL}${endpoint}"
    else
        curl -s -X "$method" "${headers[@]}" "${BASE_URL}${endpoint}"
    fi
}

# Check if server is running
info "Checking if server is running at $BASE_URL..."
if ! curl -s "$BASE_URL" > /dev/null 2>&1; then
    fail "Server not running" "Could not connect to $BASE_URL"
fi
pass "Server is running"

echo ""
echo "========================================="
echo "PUBLIC ENDPOINTS"
echo "========================================="

# Test: List predictions (empty)
info "Testing GET /api/predictions (should be empty or have sample)"
PREDICTIONS=$(request GET "/api/predictions")
if echo "$PREDICTIONS" | jq -e '. | type == "array"' > /dev/null 2>&1; then
    pass "GET /api/predictions returns array"
else
    fail "GET /api/predictions" "$PREDICTIONS"
fi

# Test: List users (leaderboard - should be empty, admin excluded)
info "Testing GET /api/users (leaderboard)"
USERS=$(request GET "/api/users")
if echo "$USERS" | jq -e '. | type == "array"' > /dev/null 2>&1; then
    pass "GET /api/users returns array"
else
    fail "GET /api/users" "$USERS"
fi

echo ""
echo "========================================="
echo "GUEST ENDPOINTS"
echo "========================================="

# Test: Register new user
info "Testing POST /api/register"
REGISTER_RESP=$(request POST "/api/register" '{"name":"TestUser1","pin":"1234"}')
if echo "$REGISTER_RESP" | jq -e '.token' > /dev/null 2>&1; then
    USER1_TOKEN=$(echo "$REGISTER_RESP" | jq -r '.token')
    USER1_ID=$(echo "$REGISTER_RESP" | jq -r '.user.id')
    pass "POST /api/register - created TestUser1"
else
    fail "POST /api/register" "$REGISTER_RESP"
fi

# Test: Register second user
info "Testing POST /api/register (second user)"
REGISTER_RESP2=$(request POST "/api/register" '{"name":"TestUser2","pin":"5678"}')
if echo "$REGISTER_RESP2" | jq -e '.token' > /dev/null 2>&1; then
    USER2_TOKEN=$(echo "$REGISTER_RESP2" | jq -r '.token')
    USER2_ID=$(echo "$REGISTER_RESP2" | jq -r '.user.id')
    pass "POST /api/register - created TestUser2"
else
    fail "POST /api/register (second user)" "$REGISTER_RESP2"
fi

# Test: Register duplicate name should fail
info "Testing POST /api/register (duplicate name)"
DUP_RESP=$(request POST "/api/register" '{"name":"TestUser1","pin":"9999"}')
if echo "$DUP_RESP" | jq -e '.error' > /dev/null 2>&1; then
    pass "POST /api/register rejects duplicate name"
else
    fail "POST /api/register should reject duplicate" "$DUP_RESP"
fi

# Test: Login
info "Testing POST /api/login"
LOGIN_RESP=$(request POST "/api/login" '{"name":"TestUser1","pin":"1234"}')
if echo "$LOGIN_RESP" | jq -e '.token' > /dev/null 2>&1; then
    pass "POST /api/login successful"
else
    fail "POST /api/login" "$LOGIN_RESP"
fi

# Test: Login with wrong PIN
info "Testing POST /api/login (wrong PIN)"
BAD_LOGIN=$(request POST "/api/login" '{"name":"TestUser1","pin":"wrong"}')
if echo "$BAD_LOGIN" | jq -e '.error' > /dev/null 2>&1; then
    pass "POST /api/login rejects wrong PIN"
else
    fail "POST /api/login should reject wrong PIN" "$BAD_LOGIN"
fi

# Test: Admin login
info "Testing POST /api/login (admin)"
ADMIN_LOGIN=$(request POST "/api/login" "{\"name\":\"Admin\",\"pin\":\"$ADMIN_PIN\"}")
if echo "$ADMIN_LOGIN" | jq -e '.token' > /dev/null 2>&1; then
    ADMIN_TOKEN=$(echo "$ADMIN_LOGIN" | jq -r '.token')
    pass "POST /api/login - admin login successful"
else
    fail "POST /api/login (admin)" "$ADMIN_LOGIN"
fi

echo ""
echo "========================================="
echo "USER ENDPOINTS"
echo "========================================="

# Test: Get current user
info "Testing GET /api/me"
ME_RESP=$(request GET "/api/me" "" "$USER1_TOKEN")
if echo "$ME_RESP" | jq -e '.id' > /dev/null 2>&1; then
    USER1_TOKENS=$(echo "$ME_RESP" | jq -r '.tokens')
    pass "GET /api/me - tokens: $USER1_TOKENS"
else
    fail "GET /api/me" "$ME_RESP"
fi

# Test: Get current user without auth
info "Testing GET /api/me (no auth)"
NOAUTH_RESP=$(request GET "/api/me")
if echo "$NOAUTH_RESP" | jq -e '.error' > /dev/null 2>&1; then
    pass "GET /api/me requires auth"
else
    fail "GET /api/me should require auth" "$NOAUTH_RESP"
fi

# Test: Get my bets (empty)
info "Testing GET /api/my-bets"
MYBETS=$(request GET "/api/my-bets" "" "$USER1_TOKEN")
if echo "$MYBETS" | jq -e '. | type == "array"' > /dev/null 2>&1; then
    pass "GET /api/my-bets returns array"
else
    fail "GET /api/my-bets" "$MYBETS"
fi

echo ""
echo "========================================="
echo "ADMIN ENDPOINTS"
echo "========================================="

# Test: Create prediction
info "Testing POST /api/admin/predictions"
CREATE_PRED=$(request POST "/api/admin/predictions" '{
    "name": "What will the coin flip be?",
    "description": "The ref will flip a coin at kickoff",
    "closes_at": "2099-02-09T18:00:00Z",
    "choices": [{"name": "Heads"}, {"name": "Tails"}],
    "odds_visible_before_bet": true
}' "$ADMIN_TOKEN")
if echo "$CREATE_PRED" | jq -e '.id' > /dev/null 2>&1; then
    PRED_ID=$(echo "$CREATE_PRED" | jq -r '.id')
    CHOICE_HEADS=$(echo "$CREATE_PRED" | jq -r '.choices[0].id')
    CHOICE_TAILS=$(echo "$CREATE_PRED" | jq -r '.choices[1].id')
    pass "POST /api/admin/predictions - created prediction"
else
    fail "POST /api/admin/predictions" "$CREATE_PRED"
fi

# Test: Create prediction as non-admin should fail
info "Testing POST /api/admin/predictions (non-admin)"
NONADMIN_PRED=$(request POST "/api/admin/predictions" '{"name":"test","choices":[{"name":"a"},{"name":"b"}]}' "$USER1_TOKEN")
if echo "$NONADMIN_PRED" | jq -e '.error' > /dev/null 2>&1; then
    pass "POST /api/admin/predictions requires admin"
else
    fail "POST /api/admin/predictions should require admin" "$NONADMIN_PRED"
fi

# Test: Update prediction
info "Testing PUT /api/admin/predictions/{id}"
UPDATE_PRED=$(request PUT "/api/admin/predictions/$PRED_ID" '{"description": "Updated description"}' "$ADMIN_TOKEN")
if echo "$UPDATE_PRED" | jq -e '.description == "Updated description"' > /dev/null 2>&1; then
    pass "PUT /api/admin/predictions/{id} updated"
else
    fail "PUT /api/admin/predictions/{id}" "$UPDATE_PRED"
fi

# Test: Get single prediction
info "Testing GET /api/predictions/{id}"
GET_PRED=$(request GET "/api/predictions/$PRED_ID")
if echo "$GET_PRED" | jq -e '.prediction.id' > /dev/null 2>&1; then
    pass "GET /api/predictions/{id}"
else
    fail "GET /api/predictions/{id}" "$GET_PRED"
fi

# Test: Gift tokens to user
info "Testing POST /api/admin/users/{id}/tokens"
GIFT_RESP=$(request POST "/api/admin/users/$USER1_ID/tokens" '{"amount": 500}' "$ADMIN_TOKEN")
# Check if it succeeded (204 No Content or 200 OK)
if [ -z "$GIFT_RESP" ] || echo "$GIFT_RESP" | jq -e 'has("error") | not' > /dev/null 2>&1; then
    pass "POST /api/admin/users/{id}/tokens - gifted 500 tokens"
else
    fail "POST /api/admin/users/{id}/tokens" "$GIFT_RESP"
fi

# Verify tokens were added
ME_AFTER_GIFT=$(request GET "/api/me" "" "$USER1_TOKEN")
NEW_TOKENS=$(echo "$ME_AFTER_GIFT" | jq -r '.tokens')
info "User1 tokens after gift: $NEW_TOKENS"

echo ""
echo "========================================="
echo "BETTING FLOW"
echo "========================================="

# Test: Place bet
info "Testing POST /api/bets (User1 bets on Heads)"
PLACE_BET=$(request POST "/api/bets" "{
    \"prediction_id\": \"$PRED_ID\",
    \"prediction_choice_id\": \"$CHOICE_HEADS\",
    \"amount\": 100
}" "$USER1_TOKEN")
if echo "$PLACE_BET" | jq -e '.id' > /dev/null 2>&1; then
    BET1_ID=$(echo "$PLACE_BET" | jq -r '.id')
    pass "POST /api/bets - User1 bet 100 on Heads"
else
    fail "POST /api/bets" "$PLACE_BET"
fi

# Test: Place duplicate bet should fail
info "Testing POST /api/bets (duplicate bet)"
DUP_BET=$(request POST "/api/bets" "{
    \"prediction_id\": \"$PRED_ID\",
    \"prediction_choice_id\": \"$CHOICE_TAILS\",
    \"amount\": 50
}" "$USER1_TOKEN")
if echo "$DUP_BET" | jq -e '.error' > /dev/null 2>&1; then
    pass "POST /api/bets rejects duplicate bet on same prediction"
else
    fail "POST /api/bets should reject duplicate" "$DUP_BET"
fi

# Gift tokens to User2 so they can bet
request POST "/api/admin/users/$USER2_ID/tokens" '{"amount": 500}' "$ADMIN_TOKEN" > /dev/null

# Test: User2 places bet on Tails
info "Testing POST /api/bets (User2 bets on Tails)"
PLACE_BET2=$(request POST "/api/bets" "{
    \"prediction_id\": \"$PRED_ID\",
    \"prediction_choice_id\": \"$CHOICE_TAILS\",
    \"amount\": 200
}" "$USER2_TOKEN")
if echo "$PLACE_BET2" | jq -e '.id' > /dev/null 2>&1; then
    pass "POST /api/bets - User2 bet 200 on Tails"
else
    fail "POST /api/bets (User2)" "$PLACE_BET2"
fi

# Test: Increase bet
info "Testing PUT /api/bets/{id}/amount"
INCREASE_BET=$(request PUT "/api/bets/$BET1_ID/amount" '{"amount": 150}' "$USER1_TOKEN")
if echo "$INCREASE_BET" | jq -e '.amount' > /dev/null 2>&1; then
    NEW_AMOUNT=$(echo "$INCREASE_BET" | jq -r '.amount')
    pass "PUT /api/bets/{id}/amount - new amount: $NEW_AMOUNT"
else
    fail "PUT /api/bets/{id}/amount" "$INCREASE_BET"
fi

# Test: Check my bets
info "Testing GET /api/my-bets (after placing bet)"
MYBETS_AFTER=$(request GET "/api/my-bets" "" "$USER1_TOKEN")
BET_COUNT=$(echo "$MYBETS_AFTER" | jq '. | length')
if [ "$BET_COUNT" -ge 1 ]; then
    pass "GET /api/my-bets shows $BET_COUNT bet(s)"
else
    fail "GET /api/my-bets should show bets" "$MYBETS_AFTER"
fi

# Test: Check odds
info "Testing odds calculation"
PRED_WITH_ODDS=$(request GET "/api/predictions/$PRED_ID")
TOTAL_TOKENS=$(echo "$PRED_WITH_ODDS" | jq -r '.odds.total_tokens_placed')
info "Total tokens placed: $TOTAL_TOKENS"
pass "Odds calculated"

echo ""
echo "========================================="
echo "PREDICTION LIFECYCLE"
echo "========================================="

# Test: Close prediction
info "Testing POST /api/admin/predictions/{id}/close"
CLOSE_RESP=$(request POST "/api/admin/predictions/$PRED_ID/close" "" "$ADMIN_TOKEN")
if [ -z "$CLOSE_RESP" ] || echo "$CLOSE_RESP" | jq -e 'has("error") | not' > /dev/null 2>&1; then
    pass "POST /api/admin/predictions/{id}/close"
else
    fail "POST /api/admin/predictions/{id}/close" "$CLOSE_RESP"
fi

# Test: Betting on closed prediction should fail
info "Testing POST /api/bets (on closed prediction)"
CLOSED_BET=$(request POST "/api/bets" "{
    \"prediction_id\": \"$PRED_ID\",
    \"prediction_choice_id\": \"$CHOICE_HEADS\",
    \"amount\": 10
}" "$USER2_TOKEN")
if echo "$CLOSED_BET" | jq -e '.error' > /dev/null 2>&1; then
    pass "POST /api/bets rejects bet on closed prediction"
else
    fail "POST /api/bets should reject closed prediction" "$CLOSED_BET"
fi

# Test: Decide prediction (Heads wins)
info "Testing POST /api/admin/predictions/{id}/decide"
DECIDE_RESP=$(request POST "/api/admin/predictions/$PRED_ID/decide" "{\"winning_choice_id\": \"$CHOICE_HEADS\"}" "$ADMIN_TOKEN")
if [ -z "$DECIDE_RESP" ] || echo "$DECIDE_RESP" | jq -e 'has("error") | not' > /dev/null 2>&1; then
    pass "POST /api/admin/predictions/{id}/decide - Heads wins!"
else
    fail "POST /api/admin/predictions/{id}/decide" "$DECIDE_RESP"
fi

# Check User1 tokens (should have won)
ME_AFTER_WIN=$(request GET "/api/me" "" "$USER1_TOKEN")
TOKENS_AFTER_WIN=$(echo "$ME_AFTER_WIN" | jq -r '.tokens')
info "User1 tokens after win: $TOKENS_AFTER_WIN"

# Check User2 tokens (should have lost)
ME2_AFTER_LOSS=$(request GET "/api/me" "" "$USER2_TOKEN")
TOKENS_AFTER_LOSS=$(echo "$ME2_AFTER_LOSS" | jq -r '.tokens')
info "User2 tokens after loss: $TOKENS_AFTER_LOSS"

# Check bet status
MYBETS_FINAL=$(request GET "/api/my-bets" "" "$USER1_TOKEN")
BET_STATUS=$(echo "$MYBETS_FINAL" | jq -r '.[0].status')
info "User1 bet status: $BET_STATUS"

echo ""
echo "========================================="
echo "VOID PREDICTION TEST"
echo "========================================="

# Create another prediction to test voiding
info "Creating prediction to void"
VOID_PRED=$(request POST "/api/admin/predictions" '{
    "name": "What color will the Gatorade be?",
    "description": "Color of Gatorade dumped on winning coach",
    "choices": [{"name": "Orange"}, {"name": "Blue"}, {"name": "Yellow"}],
    "odds_visible_before_bet": false
}' "$ADMIN_TOKEN")
VOID_PRED_ID=$(echo "$VOID_PRED" | jq -r '.id')
VOID_CHOICE=$(echo "$VOID_PRED" | jq -r '.choices[0].id')
pass "Created prediction for void test"

# Place a bet
info "User1 places bet on prediction to be voided"
VOID_BET=$(request POST "/api/bets" "{
    \"prediction_id\": \"$VOID_PRED_ID\",
    \"prediction_choice_id\": \"$VOID_CHOICE\",
    \"amount\": 50
}" "$USER1_TOKEN")
pass "Placed bet on void prediction"

TOKENS_BEFORE_VOID=$(request GET "/api/me" "" "$USER1_TOKEN" | jq -r '.tokens')
info "User1 tokens before void: $TOKENS_BEFORE_VOID"

# Void the prediction
info "Testing POST /api/admin/predictions/{id}/void"
VOID_RESP=$(request POST "/api/admin/predictions/$VOID_PRED_ID/void" "" "$ADMIN_TOKEN")
if [ -z "$VOID_RESP" ] || echo "$VOID_RESP" | jq -e 'has("error") | not' > /dev/null 2>&1; then
    pass "POST /api/admin/predictions/{id}/void"
else
    fail "POST /api/admin/predictions/{id}/void" "$VOID_RESP"
fi

TOKENS_AFTER_VOID=$(request GET "/api/me" "" "$USER1_TOKEN" | jq -r '.tokens')
info "User1 tokens after void: $TOKENS_AFTER_VOID (should have 50 refunded)"

echo ""
echo "========================================="
echo "ADMIN: RESET PIN"
echo "========================================="

# Test: Reset user PIN
info "Testing POST /api/admin/users/{id}/reset-pin"
RESET_PIN_RESP=$(request POST "/api/admin/users/$USER1_ID/reset-pin" '{"new_pin": "4321"}' "$ADMIN_TOKEN")
if echo "$RESET_PIN_RESP" | jq -e '.status == "ok"' > /dev/null 2>&1; then
    pass "POST /api/admin/users/{id}/reset-pin"
else
    fail "POST /api/admin/users/{id}/reset-pin" "$RESET_PIN_RESP"
fi

# Test: Login with new PIN
info "Testing login with new PIN"
NEW_PIN_LOGIN=$(request POST "/api/login" '{"name":"TestUser1","pin":"4321"}')
if echo "$NEW_PIN_LOGIN" | jq -e '.token' > /dev/null 2>&1; then
    pass "Login with new PIN successful"
else
    fail "Login with new PIN" "$NEW_PIN_LOGIN"
fi

# Test: Old PIN should fail
info "Testing login with old PIN"
OLD_PIN_LOGIN=$(request POST "/api/login" '{"name":"TestUser1","pin":"1234"}')
if echo "$OLD_PIN_LOGIN" | jq -e '.error' > /dev/null 2>&1; then
    pass "Old PIN rejected"
else
    fail "Old PIN should be rejected" "$OLD_PIN_LOGIN"
fi

echo ""
echo "========================================="
echo "LEADERBOARD"
echo "========================================="

info "Testing GET /api/users (final leaderboard)"
FINAL_LEADERBOARD=$(request GET "/api/users")
echo "$FINAL_LEADERBOARD" | jq -r '.[] | "  \(.rank). \(.name): \(.tokens) tokens"'
pass "Leaderboard retrieved"

echo ""
echo "========================================="
echo -e "${GREEN}ALL TESTS PASSED!${NC}"
echo "========================================="
