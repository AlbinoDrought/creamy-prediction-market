#!/usr/bin/env bash
# Load test harness for Creamy Prediction Market
# Usage: ./loadtest.sh [--users N] [--duration SECS] [--url URL] [--admin-pin PIN]
set -euo pipefail

USERS=20
DURATION=120
BASE_URL="http://localhost:3000"
ADMIN_PIN="1234"

while [[ $# -gt 0 ]]; do
  case $1 in
    -n|--users)     USERS="$2";     shift 2 ;;
    -d|--duration)  DURATION="$2";  shift 2 ;;
    -u|--url)       BASE_URL="$2";  shift 2 ;;
    -p|--admin-pin) ADMIN_PIN="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: $0 [--users N] [--duration SECS] [--url URL] [--admin-pin PIN]"
      exit 0 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
STATS_DIR=$(mktemp -d)
CHILD_PIDS=()

cleanup() {
  echo ""
  echo "Stopping..."
  for pid in "${CHILD_PIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  wait 2>/dev/null || true

  # Aggregate stats
  echo ""
  echo "=== Results ==="
  local total_req=0 total_err=0
  for f in "$STATS_DIR"/*.stats; do
    [ -f "$f" ] || continue
    local req err
    req=$(grep "requests" "$f" | cut -d= -f2)
    err=$(grep "errors" "$f" | cut -d= -f2)
    total_req=$((total_req + req))
    total_err=$((total_err + err))
  done

  echo "Total requests:  $total_req"
  echo "Total errors:    $total_err"
  if [ "$DURATION" -gt 0 ]; then
    echo "Requests/sec:    $(echo "scale=1; $total_req / $DURATION" | bc)"
  fi
  if [ "$total_req" -gt 0 ]; then
    echo "Error rate:      $(echo "scale=1; $total_err * 100 / $total_req" | bc)%"
  fi

  rm -rf "$STATS_DIR"
  echo ""
  echo "Done!"
}
trap cleanup EXIT

echo "=== Creamy Prediction Market Load Test ==="
echo "Users: $USERS | Duration: ${DURATION}s | URL: $BASE_URL"
echo ""

# Check dependencies
for cmd in curl jq bc; do
  command -v "$cmd" &>/dev/null || { echo "Error: $cmd is required"; exit 1; }
done

# Health check
HTTP_CODE=$(curl -sf -o /dev/null -w '%{http_code}' "${BASE_URL}/api/predictions" 2>/dev/null || echo "000")
if [ "$HTTP_CODE" = "000" ]; then
  echo "Error: Cannot connect to $BASE_URL"
  exit 1
fi
echo "Server is reachable"

# Register admin
ADMIN_RESP=$(curl -s -X POST "${BASE_URL}/api/register" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"Admin\",\"pin\":\"${ADMIN_PIN}\"}" 2>/dev/null)
ADMIN_TOKEN=$(echo "$ADMIN_RESP" | jq -r '.token // empty' 2>/dev/null)

if [ -z "$ADMIN_TOKEN" ]; then
  ADMIN_RESP=$(curl -sf -X POST "${BASE_URL}/api/login" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"Admin\",\"pin\":\"${ADMIN_PIN}\"}" 2>/dev/null)
  ADMIN_TOKEN=$(echo "$ADMIN_RESP" | jq -r '.token // empty' 2>/dev/null)
fi

if [ -z "$ADMIN_TOKEN" ]; then
  echo "Error: Could not authenticate admin (check --admin-pin)"
  exit 1
fi
echo "Admin authenticated"

# Create initial test predictions
CLOSES=$(date -u -d "+2 hours" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+2H +%Y-%m-%dT%H:%M:%SZ)
QUESTIONS=("Will it rain tomorrow" "Best pizza topping" "Cats or dogs" "Pineapple on pizza" "Coffee or tea" "Beach or mountains" "Star Wars or Star Trek" "Cake or pie" "Hot dog is a sandwich" "Tabs or spaces")

for i in $(seq 0 9); do
  curl -sf -X POST "${BASE_URL}/api/admin/predictions" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H 'Content-Type: application/json' \
    -d "{
      \"name\": \"${QUESTIONS[$i]}?\",
      \"description\": \"Load test prediction\",
      \"closes_at\": \"${CLOSES}\",
      \"choices\": [{\"name\":\"Yes\"},{\"name\":\"No\"},{\"name\":\"Maybe\"}],
      \"odds_visible_before_bet\": true
    }" > /dev/null 2>/dev/null || true
done
echo "Created 10 test predictions"

# Spawn user simulations
echo ""
echo "Spawning $USERS users..."
for i in $(seq 1 "$USERS"); do
  bash "${SCRIPT_DIR}/loadtest-user.sh" "$BASE_URL" "$i" "$DURATION" "$STATS_DIR" &
  CHILD_PIDS+=($!)
  sleep 0.1  # stagger starts
done
echo "All users active"

# Admin loop: periodically close/decide predictions and create new ones
# This keeps tokens flowing back to users so they can keep betting
(
  COUNTER=10
  sleep 20  # let users place initial bets
  END_TIME=$(( $(date +%s) + DURATION - 10 ))

  while [ "$(date +%s)" -lt "$END_TIME" ]; do
    sleep $(( RANDOM % 15 + 10 ))

    # Pick a random open prediction, close it, decide it
    PREDS=$(curl -sf -H "Authorization: Bearer $ADMIN_TOKEN" "${BASE_URL}/api/predictions" 2>/dev/null)
    OPEN_ID=$(echo "$PREDS" | jq -r '[.[] | select(.prediction.status == "open")] | .[0].prediction.id // empty' 2>/dev/null)

    if [ -n "$OPEN_ID" ]; then
      # Close
      curl -sf -X POST "${BASE_URL}/api/admin/predictions/${OPEN_ID}/close" \
        -H "Authorization: Bearer $ADMIN_TOKEN" > /dev/null 2>/dev/null || true
      sleep 2

      # Decide (pick first choice as winner)
      PRED_DETAIL=$(curl -sf "${BASE_URL}/api/predictions/${OPEN_ID}" 2>/dev/null)
      WINNER=$(echo "$PRED_DETAIL" | jq -r '.prediction.choices[0].id // empty' 2>/dev/null)
      if [ -n "$WINNER" ]; then
        curl -sf -X POST "${BASE_URL}/api/admin/predictions/${OPEN_ID}/decide" \
          -H "Authorization: Bearer $ADMIN_TOKEN" \
          -H 'Content-Type: application/json' \
          -d "{\"winning_choice_id\":\"$WINNER\"}" > /dev/null 2>/dev/null || true
      fi

      # Create replacement
      COUNTER=$((COUNTER + 1))
      CLOSES=$(date -u -d "+2 hours" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+2H +%Y-%m-%dT%H:%M:%SZ)
      curl -sf -X POST "${BASE_URL}/api/admin/predictions" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H 'Content-Type: application/json' \
        -d "{
          \"name\": \"Load Test Question #${COUNTER}?\",
          \"description\": \"Load test prediction\",
          \"closes_at\": \"${CLOSES}\",
          \"choices\": [{\"name\":\"Yes\"},{\"name\":\"No\"},{\"name\":\"Maybe\"}],
          \"odds_visible_before_bet\": true
        }" > /dev/null 2>/dev/null || true
    fi
  done
) &
CHILD_PIDS+=($!)

echo ""
echo "Running for ${DURATION}s... (Ctrl+C to stop early)"
echo "Monitor CPU: htop, or: watch -n1 'ps -p \$(pgrep creamy-prediction) -o %cpu,%mem,rss'"
echo ""

# Wait for user scripts to finish
for pid in "${CHILD_PIDS[@]}"; do
  wait "$pid" 2>/dev/null || true
done
