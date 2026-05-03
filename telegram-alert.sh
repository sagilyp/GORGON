#!/bin/bash
if [ -f "$(dirname "$0")/.env" ]; then
    source "$(dirname "$0")/.env"
else
    echo "Ошибка: файл .env не найден!"
    exit 1
fi

MESSAGE="Обнаружена блокировка TOR: $(echo "$@")"

curl -s -X POST \
  -H 'Content-Type: application/json' \
  -d "{\"chat_id\": \"$TG_CHAT_ID\", \"text\": \"$MESSAGE\"}" \
  "https://api.telegram.org/bot$TG_BOT_TOKEN/sendMessage"


