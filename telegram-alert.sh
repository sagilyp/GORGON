#!/bin/bash
TELEGRAM_BOT_TOKEN="7779836915:AAGZJ8BaJ6se0ryjW9_KHL3INBLi8RGueRo"
CHAT_ID="-4787521880"
MESSAGE="Обнаружена блокировка TOR: $(echo "$@")"

curl -s -X POST \
  -H 'Content-Type: application/json' \
  -d "{\"chat_id\": \"$CHAT_ID\", \"text\": \"$MESSAGE\"}" \
  "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/sendMessage"


