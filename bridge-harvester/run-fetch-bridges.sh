#!/bin/bash

for i in $(seq 1 1000); do
  echo "Iteration $i/1000"
  python3 /home/sagilyp/GORGONA/bridge-harvester/fetch_bridges.py
  sleep $((RANDOM % 11 + 5)) # Случайная задержка от 5 до 15 секунд
done
