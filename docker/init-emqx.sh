#!/bin/sh

# 等待 EMQX 启动完成
until /opt/emqx/bin/emqx_ctl status >/dev/null 2>&1; do
  echo "Waiting for EMQX to start..."
  sleep 5
done

echo "EMQX is ready. Creating user..."

# 创建用户：root / root123
/opt/emqx/bin/emqx_ctl users add root root123

echo "User 'root' created successfully."