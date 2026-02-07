#!/usr/bin/env bash
# Single simulated user for load testing. Launched by loadtest.sh.
set -uo pipefail

BASE_URL="$1"
USER_NUM="$2"
DURATION="$3"
STATS_DIR="$4"

NAME="lt$(printf '%02d' "$USER_NUM")x$(( RANDOM % 9000 + 1000 ))"
PIN="0000"
AUTH=""
END_TIME=$(( $(date +%s) + DURATION ))

REQUESTS=0
ERRORS=0
SSE_PID=""

# State
PREDICTIONS_CACHE=""
SHOP_CACHE=""
BET_PREDICTION_IDS=""
BET_IDS=""
OWNED_ITEM_IDS=""
MY_TOKENS=0
MY_COINS=0

log() { echo "[user$(printf '%02d' "$USER_NUM")] $*"; }

# --- HTTP helpers ---

get() {
  REQUESTS=$((REQUESTS + 1))
  curl -s -H "Authorization: Bearer $AUTH" "${BASE_URL}$1" 2>/dev/null || { ERRORS=$((ERRORS+1)); return 1; }
}

post() {
  REQUESTS=$((REQUESTS + 1))
  if [ -n "${2:-}" ]; then
    curl -s -X POST -H "Authorization: Bearer $AUTH" -H 'Content-Type: application/json' -d "$2" "${BASE_URL}$1" 2>/dev/null
  else
    curl -s -X POST -H "Authorization: Bearer $AUTH" "${BASE_URL}$1" 2>/dev/null
  fi || { ERRORS=$((ERRORS+1)); return 1; }
}

put() {
  REQUESTS=$((REQUESTS + 1))
  if [ -n "${2:-}" ]; then
    curl -s -X PUT -H "Authorization: Bearer $AUTH" -H 'Content-Type: application/json' -d "$2" "${BASE_URL}$1" 2>/dev/null
  else
    curl -s -X PUT -H "Authorization: Bearer $AUTH" "${BASE_URL}$1" 2>/dev/null
  fi || { ERRORS=$((ERRORS+1)); return 1; }
}

# --- Lifecycle ---

register() {
  REQUESTS=$((REQUESTS + 1))
  local resp
  resp=$(curl -s -X POST "${BASE_URL}/api/register" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"$NAME\",\"pin\":\"$PIN\"}" 2>/dev/null)
  AUTH=$(echo "$resp" | jq -r '.token // empty' 2>/dev/null)
  if [ -z "$AUTH" ]; then
    ERRORS=$((ERRORS + 1))
    log "Register failed"
    return 1
  fi
  MY_TOKENS=$(echo "$resp" | jq '.user.tokens // 0' 2>/dev/null)
  MY_COINS=$(echo "$resp" | jq '.user.coins // 0' 2>/dev/null)
  log "Registered as $NAME (${MY_TOKENS} tokens, ${MY_COINS} coins)"
}

start_sse() {
  # Reactive SSE: parse events and fetch data just like the real UI does.
  # On "predictions" -> GET /api/predictions
  # On "leaderboard" -> GET /api/me + GET /api/leaderboard
  # On "bets"        -> GET /api/my-bets
  (
    while [ "$(date +%s)" -lt "$END_TIME" ]; do
      curl -s -N "${BASE_URL}/api/events?token=${AUTH}" 2>/dev/null | \
        while IFS= read -r line; do
          # SSE data lines look like: data: {"type":"predictions",...}
          case "$line" in
            data:*predictions*)
              curl -s -H "Authorization: Bearer $AUTH" "${BASE_URL}/api/predictions" > /dev/null 2>&1 &
              ;;
            data:*leaderboard*)
              curl -s -H "Authorization: Bearer $AUTH" "${BASE_URL}/api/me" > /dev/null 2>&1 &
              curl -s -H "Authorization: Bearer $AUTH" "${BASE_URL}/api/leaderboard" > /dev/null 2>&1 &
              ;;
            data:*\"bets\"*)
              curl -s -H "Authorization: Bearer $AUTH" "${BASE_URL}/api/my-bets" > /dev/null 2>&1 &
              ;;
          esac
        done || true
      sleep 2  # reconnect delay
    done
  ) &
  SSE_PID=$!
}

cleanup() {
  [ -n "$SSE_PID" ] && kill "$SSE_PID" 2>/dev/null || true
  wait "$SSE_PID" 2>/dev/null || true
  cat > "${STATS_DIR}/user${USER_NUM}.stats" <<EOF
requests=$REQUESTS
errors=$ERRORS
EOF
  log "Done: ${REQUESTS} requests, ${ERRORS} errors"
}
trap cleanup EXIT

# --- Actions ---

action_browse_predictions() {
  PREDICTIONS_CACHE=$(get "/api/predictions") || true
}

action_place_bet() {
  [ -z "$PREDICTIONS_CACHE" ] && action_browse_predictions

  local open_ids
  open_ids=$(echo "$PREDICTIONS_CACHE" | jq -r '.[] | select(.prediction.status == "open") | .prediction.id' 2>/dev/null) || return
  [ -z "$open_ids" ] && return
  [ "$MY_TOKENS" -lt 10 ] && return

  # Find a prediction we haven't bet on yet
  local target_id=""
  for pid in $open_ids; do
    case " $BET_PREDICTION_IDS " in
      *" $pid "*) continue ;;
    esac
    target_id="$pid"
    break
  done
  [ -z "$target_id" ] && return

  # Pick a random choice
  local num_choices
  num_choices=$(echo "$PREDICTIONS_CACHE" | jq "[.[] | select(.prediction.id == \"$target_id\")] | .[0].prediction.choices | length" 2>/dev/null) || return
  [ "${num_choices:-0}" -le 0 ] 2>/dev/null && return
  local choice_idx=$(( RANDOM % num_choices ))
  local choice_id
  choice_id=$(echo "$PREDICTIONS_CACHE" | jq -r "[.[] | select(.prediction.id == \"$target_id\")] | .[0].prediction.choices[$choice_idx].id" 2>/dev/null) || return
  [ -z "$choice_id" ] && return

  local amount=$(( RANDOM % 91 + 10 ))
  [ "$amount" -gt "$MY_TOKENS" ] && amount="$MY_TOKENS"

  local resp
  resp=$(post "/api/bets" "{\"prediction_id\":\"$target_id\",\"prediction_choice_id\":\"$choice_id\",\"amount\":$amount}")
  local bet_id
  bet_id=$(echo "$resp" | jq -r '.id // empty' 2>/dev/null)
  if [ -n "$bet_id" ]; then
    BET_IDS="$BET_IDS $bet_id"
    BET_PREDICTION_IDS="$BET_PREDICTION_IDS $target_id"
    MY_TOKENS=$((MY_TOKENS - amount))
  fi
}

action_increase_bet() {
  [ -z "$BET_IDS" ] && return
  [ "$MY_TOKENS" -lt 10 ] && return

  local ids=($BET_IDS)
  local idx=$(( RANDOM % ${#ids[@]} ))
  local bet_id="${ids[$idx]}"

  # Send a random higher amount (may 409 if not higher than current, that's fine)
  local new_amount=$(( RANDOM % 300 + 50 ))
  put "/api/bets/${bet_id}/amount" "{\"amount\":$new_amount}" > /dev/null || true
}

action_leaderboard() {
  get "/api/leaderboard" > /dev/null || true
}

action_my_bets() {
  get "/api/my-bets" > /dev/null || true
}

action_refresh_me() {
  local resp
  resp=$(get "/api/me") || return
  MY_TOKENS=$(echo "$resp" | jq '.tokens // 0' 2>/dev/null) || true
  MY_COINS=$(echo "$resp" | jq '.coins // 0' 2>/dev/null) || true
  OWNED_ITEM_IDS=$(echo "$resp" | jq -r '.owned_items // [] | .[]' 2>/dev/null | tr '\n' ' ') || true
}

action_shop() {
  [ -z "$SHOP_CACHE" ] && { SHOP_CACHE=$(get "/api/shop") || return; }

  local r=$(( RANDOM % 4 ))

  if [ $r -le 1 ] && [ "${MY_COINS:-0}" -gt 0 ]; then
    # Buy a random affordable non-locked item (50% of shop actions)
    local rand_val=$(( RANDOM ))
    local item_id
    item_id=$(echo "$SHOP_CACHE" | jq -r \
      --argjson c "${MY_COINS:-0}" \
      --argjson r "$rand_val" \
      '[.[] | select(.price <= $c and .price > 0 and (.locked | not) and (.consumable | not))] | if length > 0 then .[($r % length)].id else empty end' \
      2>/dev/null)
    if [ -n "$item_id" ]; then
      post "/api/shop/buy/${item_id}" > /dev/null || true
      # Auto-equip after purchase, just like the real UI does
      put "/api/shop/equip/${item_id}" > /dev/null || true
      action_refresh_me
    fi
  elif [ $r -eq 2 ] && [ -n "$OWNED_ITEM_IDS" ]; then
    # Swap to a different owned item (25% of shop actions)
    local items=($OWNED_ITEM_IDS)
    [ ${#items[@]} -eq 0 ] && return
    local idx=$(( RANDOM % ${#items[@]} ))
    put "/api/shop/equip/${items[$idx]}" > /dev/null || true
  else
    SHOP_CACHE=$(get "/api/shop") || true
  fi
}

action_minigame() {
  local score=$(( RANDOM % 500 + 30 ))
  local resp
  resp=$(post "/api/minigame/claim" "{\"score\":$score}")
  local earned
  earned=$(echo "$resp" | jq '.coins_earned // 0' 2>/dev/null) || true
  MY_COINS=$((MY_COINS + ${earned:-0}))
}

action_spin() {
  post "/api/spin" > /dev/null || true
}

action_achievements() {
  get "/api/achievements" > /dev/null || true
  get "/api/my-achievements" > /dev/null || true
}

# --- Main ---

register || exit 1
start_sse

action_browse_predictions
action_refresh_me
sleep 0.5

while [ "$(date +%s)" -lt "$END_TIME" ]; do
  R=$(( RANDOM % 100 ))

  if   [ $R -lt 15 ]; then action_browse_predictions
  elif [ $R -lt 30 ]; then action_place_bet
  elif [ $R -lt 38 ]; then action_increase_bet
  elif [ $R -lt 50 ]; then action_leaderboard
  elif [ $R -lt 58 ]; then action_my_bets
  elif [ $R -lt 66 ]; then action_refresh_me
  elif [ $R -lt 76 ]; then action_shop
  elif [ $R -lt 88 ]; then action_minigame
  elif [ $R -lt 94 ]; then action_spin
  else                      action_achievements
  fi

  # 1.0-3.9s between actions (human think time)
  sleep "$(( RANDOM % 3 + 1 )).$(( RANDOM % 10 ))"
done
