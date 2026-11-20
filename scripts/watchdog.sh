#!/usr/bin/env bash
# 自愈守护：每 15 秒检查服务健康，连续 3 次失败则自动重启，并写告警日志。
# 由 deploy.sh 拉起，独立后台运行。
set -uo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="/data/gobin"
RUN_DIR="/data/gobin/run"
LOG_DIR="/data/gobin/logs"
CURRENT_LINK="$BIN_DIR/server"
PID_FILE="$RUN_DIR/server.pid"
PORT="${APP_PORT:-8080}"
HEALTH_URL="http://127.0.0.1:${PORT}/healthz"
ALERT_LOG="$LOG_DIR/alerts.log"
CHECK_INTERVAL="${WATCHDOG_INTERVAL:-15}"
MAX_FAILS=3

mkdir -p "$LOG_DIR" "$RUN_DIR"
log() { echo "[$(date -u +%FT%TZ)] $*" | tee -a "$ALERT_LOG"; }

log "自愈守护启动 (间隔=${CHECK_INTERVAL}s, 阈值=${MAX_FAILS}次)"
fails=0

restart_service() {
  log "告警: 服务连续 ${fails} 次健康检查失败，尝试自动重启"
  if [ -f "$PID_FILE" ]; then
    old_pid="$(cat "$PID_FILE")"
    kill -TERM "$old_pid" 2>/dev/null || true
    sleep 2
    kill -KILL "$old_pid" 2>/dev/null || true
  fi
  if [ -x "$CURRENT_LINK" ]; then
    APP_VERSION="${APP_VERSION:-recovered}" nohup "$CURRENT_LINK" >>"$LOG_DIR/server.log" 2>&1 &
    echo $! > "$PID_FILE"
    log "服务已自动拉起, 新 PID=$(cat "$PID_FILE")"
  else
    log "错误: 找不到可执行文件 $CURRENT_LINK，无法重启"
  fi
}

while true; do
  if curl -fsS --max-time 3 "$HEALTH_URL" >/dev/null 2>&1; then
    if [ "$fails" -gt 0 ]; then
      log "服务已恢复健康"
    fi
    fails=0
  else
    fails=$((fails + 1))
    log "健康检查失败 ($fails/$MAX_FAILS)"
    if [ "$fails" -ge "$MAX_FAILS" ]; then
      restart_service
      fails=0
      sleep 5
    fi
  fi
  sleep "$CHECK_INTERVAL"
done
