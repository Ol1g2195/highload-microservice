#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
echo "Smoke start => ${BASE_URL}"

code=$(curl -s -o /dev/null -w '%{http_code}' "${BASE_URL}/health")
if [ "$code" != "200" ]; then echo "Health failed: $code" >&2; exit 1; fi

user_payload='{"email":"user@example.com","first_name":"John","last_name":"Doe"}'
created=$(curl -s -X POST "${BASE_URL}/api/v1/users" -H 'Content-Type: application/json' -d "$user_payload")
id=$(echo "$created" | jq -r '.id')
if [ -z "$id" ] || [ "$id" = "null" ]; then echo "User create failed" >&2; exit 1; fi
echo "User: $id"

curl -s "${BASE_URL}/api/v1/users/${id}" >/dev/null
curl -s "${BASE_URL}/api/v1/users/${id}" >/dev/null

event_payload="{\"user_id\":\"$id\",\"type\":\"smoke_event\",\"data\":\"{\\\"action\\\":\\\"login\\\"}\"}"
curl -s -X POST "${BASE_URL}/api/v1/events" -H 'Content-Type: application/json' -d "$event_payload" >/dev/null

events=$(curl -s "${BASE_URL}/api/v1/events?page=1&limit=10")
echo "$events" | jq . >/dev/null

echo "Smoke OK"


