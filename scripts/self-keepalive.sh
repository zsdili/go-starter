#!/usr/bin/env bash
# 机器内部保活：每 5 分钟调 API 续租，让自身保持 running。
# 租约使机器不休眠 -> 机器 running 本循环才能运行 -> 继续续租，形成自维持。
set -uo pipefail
KEY="ibk_zn3dgd2orzzxyx7iokspqfq5tpazrp3j7z7lexfi5sxrpcttznpq"
MID="m_agq7uyx5ft27oh4wi2ki4rdpee"
API="https://api.instabox.sh"
LEASE="keepalive"
TTL=900
INTERVAL=300
LOG="/data/gobin/logs/self-keepalive.log"
mkdir -p "$(dirname "$LOG")"
log(){ echo "[$(date -u +%FT%TZ)] $*" >> "$LOG"; }
log "内部保活启动 interval=${INTERVAL}s ttl=${TTL}s"
while true; do
  code=$(curl -sS -m 30 -o /tmp/ka.out -w '%{http_code}' \
    -X PUT -H "X-API-Key: $KEY" -H 'Content-Type: application/json' \
    -d "{\"ttlSeconds\":$TTL}" \
    "$API/v1/machines/$MID/activity-leases/$LEASE" 2>/dev/null || echo 000)
  if [ "$code" = "200" ] || [ "$code" = "204" ]; then
    exp=$(grep -o '"expiresAt":"[^"]*"' /tmp/ka.out | cut -d'"' -f4)
    log "续租成功 HTTP $code 到期 $exp"
  else
    log "续租失败 HTTP $code: $(cat /tmp/ka.out 2>/dev/null | head -c 200)"
  fi
  sleep "$INTERVAL"
done
