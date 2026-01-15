#!/bin/bash

# 1. Avatars API -> 8000
PORT=8000 ./bin/avatars-api &

# 2. Calendars API -> 8001
PORT=8001 node ./calendars-api/dist/index.js &

# 3. Charts API -> 8002
PORT=8002 ./bin/charts-api &

# 4. GitHub Dashboard API -> 8003
PORT=8003 ./bin/github-stats &

# 5. Institutions API -> 8004
PORT=8004 ./bin/institutions-api &

# 6. Locations API -> 8005
PORT=8005 ./bin/locations-api &

# 7. NPM Dashboard API -> 8006
PORT=8006 ./bin/npm-stats &

# Start Caddy on Port 7999 (The Public Access Point)
echo "🚀 Starting Utility Hub on port 7999..."
caddy run --config Caddyfile --adapter caddyfile