#!/bin/bash
# ./setup-infra.sh
# 用途：一键部署 Yunyez 基础设施（DB + Cache + MQTT）
#

set -e  # 遇错退出

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR/.."

echo "🚀 Starting Yunyez infrastructure setup..."

# 进入项目根目录
cd "$PROJECT_ROOT"

# 赋予 EMQX 初始化脚本执行权限
chmod +x "$SCRIPT_DIR/docker/init-emqx.sh"

# 启动基础设施
docker compose -f docker/docker-compose.yml up -d

echo ""
echo "   Infrastructure started!"
echo "   PostgreSQL: localhost:5432 (user: postgres, pass: root)"
echo "   Redis:      localhost:6379 (no password)"
echo "   EMQX:       localhost:1883 (user: root, pass: root123)"
echo "   Dashboard:  http://<server-ip>:18083"
echo ""
echo "   To stop: docker compose -f docker-compose.infra.yml down"
echo "   To rebuild: re-run this script"