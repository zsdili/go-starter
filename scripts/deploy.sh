#!/usr/bin/env bash
# 部署脚本：编译 -> 版本化二进制 -> 优雅重启 -> 健康检查 -> 失败回滚
# 由 GitHub Actions CD 在机器上调用，也可手动执行：bash scripts/deploy.sh
set -euo pipefail

APP_NAME="${APP_NAME:-go-starter}"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="/data/gobin"
RUN_DIR="/data/gobin/run"
LOG_DIR="/data/gobin/logs"
CURRENT_LINK="$BIN_DIR/server"
PID_FILE="$RUN_DIR/server.pid"
PORT="${APP_PORT:-8080}"
HEALTH_URL="http://127.0.0.1:${PORT}/healthz"
VERSION="${VERSION:-$(git -C "$PROJECT_DIR" rev-parse --short HEAD 2>/dev/null || echo manual)}"
DEPLOY_LOG="$LOG_DIR/deploy.log"

mkdir -p "$BIN_DIR" "$RUN_DIR" "$LOG_DIR"
log() { echo "[$(date -u +%FT%TZ)] $*" | tee -a "$DEPLOY_LOG"; }

log "==== 开始部署 version=$VERSION ===="

# 1) 编译到版本化路径
log "编译中..."
cd "$PROJECT_DIR"
export CGO_ENABLED=0
go build -trimpath \
  -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${VERSION} -X main.buildAt=$(date -u +%FT%TZ)" \
  -o "$BIN_DIR/server.${VERSION}" ./cmd/server
log "编译完成: $BIN_DIR/server.${VERSION}"

# 记录旧版本用于回滚
OLD_TARGET=""
if [ -L "$CURRENT_LINK" ]; then
  OLD_TARGET="$(readlink "$CURRENT_LINK")"
fi

# 2) 切换软链到新版本
ln -sfn "$BIN_DIR/server.${VERSION}" "$CURRENT_LINK"
log "已切换当前版本 -> server.${VERSION}"

# 3) 优雅停止旧进程
# 先杀 PID 文件记录的进程，再兜底杀掉所有占用端口/可执行文件的残留进程
stop_old() {
  if [ -f "$PID_FILE" ]; then
    old_pid="$(cat "$PID_FILE")"
    if kill -0 "$old_pid" 2>/dev/null; then
      log "发送 SIGTERM 到旧进程 $old_pid"
      kill -TERM "$old_pid" 2>/dev/null || true
      for _ in $(seq 1 20); do
        kill -0 "$old_pid" 2>/dev/null || break
        sleep 0.5
      done
      kill -KILL "$old_pid" 2>/dev/null || true
    fi
    rm -f "$PID_FILE"
  fi
  # 兜底：杀掉所有运行本服务二进制的残留进程（含历史手动启动、未写 PID 的进程）
  for p in $(pgrep -f "$BIN_DIR/server" 2>/dev/null || true); do
    [ "$p" = "$$" ] && continue
    log "清理残留进程 $p"
    kill -TERM "$p" 2>/dev/null || true
  done
  sleep 1
  for p in $(pgrep -f "$BIN_DIR/server" 2>/dev/null || true); do
    [ "$p" = "$$" ] && continue
    kill -KILL "$p" 2>/dev/null || true
  done
  # 确认端口已释放
  for _ in $(seq 1 10); do
    if curl -fsS --max-time 1 "$HEALTH_URL" >/dev/null 2>&1; then
      sleep 0.5
    else
      break
    fi
  done
}
stop_old

# 4) 启动新进程
log "启动新进程..."
APP_VERSION="$VERSION" APP_NAME="$APP_NAME" nohup "$CURRENT_LINK" \
  >>"$LOG_DIR/server.log" 2>&1 &
new_pid=$!
echo "$new_pid" > "$PID_FILE"
log "新进程 PID=$new_pid"

# 5) 健康检查（最多等 15 秒）：必须"新进程存活"且"健康接口 200"同时满足
healthy=0
for i in $(seq 1 30); do
  if ! kill -0 "$new_pid" 2>/dev/null; then
    log "新进程 $new_pid 已退出，停止等待"
    break
  fi
  if curl -fsS --max-time 2 "$HEALTH_URL" >/dev/null 2>&1; then
    # 再校验版本，确认响应来自新部署的版本而非残留旧进程
    served_version="$(curl -fsS --max-time 2 http://127.0.0.1:${PORT}/api/info 2>/dev/null | grep -o "\"version\":\"[^\"]*\"" | cut -d'"' -f4 || true)"
    if [ "$served_version" = "$VERSION" ]; then
      healthy=1
      break
    fi
  fi
  sleep 0.5
done

# 6) 成功/回滚
if [ "$healthy" -eq 1 ]; then
  log "健康检查通过，部署成功 ✓ version=$VERSION"
  # 确保自愈守护在运行
  if ! pgrep -f "watchdog.sh" >/dev/null 2>&1; then
    nohup bash "$PROJECT_DIR/scripts/watchdog.sh" >>"$LOG_DIR/watchdog.log" 2>&1 &
    log "已启动自愈守护进程"
  fi
else
  log "健康检查失败，执行回滚..."
  kill -KILL "$new_pid" 2>/dev/null || true
  rm -f "$PID_FILE"
  if [ -n "$OLD_TARGET" ] && [ -x "$OLD_TARGET" ]; then
    ln -sfn "$OLD_TARGET" "$CURRENT_LINK"
    nohup "$CURRENT_LINK" >>"$LOG_DIR/server.log" 2>&1 &
    echo $! > "$PID_FILE"
    log "已回滚到 $OLD_TARGET"
  fi
  log "部署失败 ✗，请检查日志: $LOG_DIR/server.log"
  tail -20 "$LOG_DIR/server.log" >> "$DEPLOY_LOG" 2>/dev/null || true
  exit 1
fi
