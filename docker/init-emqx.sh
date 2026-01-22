#!/bin/bash
set -e

echo "⏳ Waiting for EMQX to be ready..."
while ! docker exec yunyez_emqx /opt/emqx/bin/emqx_ctl status >/dev/null 2>&1; do
  echo "EMQX not ready yet... waiting 5s"
  sleep 5
done

echo "✅ EMQX is up. Creating user 'root'..."
docker exec yunyez_emqx /opt/emqx/bin/emqx_ctl users add root root123

echo "🎉 User created. Dashboard: http://<IP>:18083 (root / root123)"