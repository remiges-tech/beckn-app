#!/bin/sh
# Runs automatically on every nginx container start via /docker-entrypoint.d/.
# BPP_ID / BPP_URI / NET_ID are now BAP server-side config — frontend only
# needs the API base path.
cat > /usr/share/nginx/html/env-config.js <<EOF
window.__env__ = {
  API_BASE: "${API_BASE:-/api/v1}"
};
EOF
