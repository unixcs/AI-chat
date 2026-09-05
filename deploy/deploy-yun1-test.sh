#!/bin/bash
# AI-chat Go 版 · yun1 测试实例部署/更新脚本
# 在 WSL 本地（/mnt/demo/AI-chat）执行：
#   bash deploy/deploy-yun1-test.sh
# 红线：只碰 yun1 的 ai-chat-go-* 容器与 /opt/ai-chat-go-test/，绝不触碰
#       yun 生产、yun1 现有 ai-chat 二节点（8181/3001）及其 /opt/AI-chat。

set -euo pipefail

HOST="${SSH_TARGET:-yun1}"
REMOTE_DIR=/opt/ai-chat-go-test
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> [1/5] 本地构建镜像（WSL，不在 yun1 构建）"
docker build -q -t ai-chat-go-backend:latest "$REPO_ROOT/backend-go"
docker build -q -t ai-chat-go-frontend:latest "$REPO_ROOT/frontend"

echo "==> [2/5] 准备远端目录"
ssh "$HOST" "mkdir -p $REMOTE_DIR/backend $REMOTE_DIR/frontend $REMOTE_DIR/data"

echo "==> [3/5] 初始化远端 .env（已存在则跳过；密钥从同机现有 ai-chat env 提取，不回传本地）"
ssh "$HOST" "test -f $REMOTE_DIR/backend/.env || { \
  echo 'NODE_ENV=production'; \
  echo 'PORT=3001'; \
  echo 'SQLITE_PATH=/app/data/data.sqlite'; \
  echo 'DEEPSEEK_SYSTEM_PROMPT_FILE=prompts/Prompt.md'; \
  echo 'AI_MODE=official'; \
  grep -E '^(DEEPSEEK_API_KEY|DEEPSEEK_MODEL|DEEPSEEK_BASE_URL|JWT_SECRET|DEEPSEEK_EXTRA_BODY)=' /opt/AI-chat/backend/.env; \
} > $REMOTE_DIR/backend/.env && chmod 600 $REMOTE_DIR/backend/.env && echo 'env created' || echo 'env kept'"
scp -q "$REPO_ROOT/deploy/docker-compose.yun1-test.yml" "$HOST:$REMOTE_DIR/docker-compose.yml"

echo "==> [4/5] 传输镜像并加载"
for img in ai-chat-go-backend:latest ai-chat-go-frontend:latest; do
  echo "  - $img"
  docker save "$img" | gzip -1 | ssh "$HOST" 'gunzip | docker load'
done

echo "==> [5/5] 拉起服务（只影响 ai-chat-go-* 容器）"
ssh "$HOST" "cd $REMOTE_DIR && docker compose -p ai-chat-go-test up -d --no-build"

echo "==> 验收"
sleep 2
ssh "$HOST" "curl -s --max-time 5 http://127.0.0.1:3002/api/health && echo && docker ps --filter name=ai-chat-go- --format '{{.Names}} {{.Status}} {{.Ports}}'"
