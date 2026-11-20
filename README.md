# go-starter

一个工程化、自动化、智能化的 Go HTTP 服务脚手架，部署在 InstaBox 免费云主机（wellknow-box）上。零外部依赖（仅用 Go 标准库），离线可编译。

## 特性

### 工程化
- 标准分层：`cmd`（入口）/ `internal`（config、server、handler）
- 环境变量配置 + 默认值 + 校验（`internal/config`）
- 结构化日志（JSON/text，`log/slog`）
- 优雅关闭（SIGINT/SIGTERM + `http.Server.Shutdown`）
- 中间件：RequestID、访问日志、panic 恢复
- 单元测试 + 基准测试 + 竞态检测
- Makefile、多阶段 Dockerfile、golangci-lint 配置

### 自动化（GitHub Actions）
- **CI**（`.github/workflows/ci.yml`）：每次 push/PR 自动 gofmt 检查、go vet、`go test -race`、覆盖率、5 平台交叉编译并上传产物，带 Go 模块缓存
- **CD**（`.github/workflows/cd.yml`）：push 到 main 后自动打包源码 → 经 `instabox exec --stdin` 上传 → 机器执行 `deploy.sh`（编译→版本化切换→优雅重启→健康检查→失败回滚）；支持手动触发

### 智能化（自愈运维）
- 版本化二进制 + 软链切换，天然支持回滚
- 部署后健康检查，失败自动回滚到上一版本
- `watchdog.sh` 自愈守护：每 15 秒检查健康，连续 3 次失败自动拉起服务，写告警日志
- 日志分离：`server.log`（业务）、`deploy.log`（部署）、`alerts.log`（告警）

## API

| 路径 | 说明 |
|------|------|
| `GET /` | 欢迎信息 |
| `GET /healthz` | 存活探针 |
| `GET /readyz` | 就绪探针 |
| `GET /api/info` | 运行时信息（版本、Go 版本、CPU、协程、请求数、内存）|
| `GET /metrics` | Prometheus 文本格式指标 |

## 配置（环境变量）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `APP_ADDR` | `:8080` | 监听地址 |
| `APP_NAME` | `go-starter` | 应用名 |
| `APP_VERSION` | `dev` | 版本 |
| `LOG_LEVEL` | `info` | debug/info/warn/error |
| `LOG_FORMAT` | `json` | json/text |
| `APP_READ_TIMEOUT` | `5s` | 读超时 |
| `APP_WRITE_TIMEOUT` | `10s` | 写超时 |
| `APP_SHUTDOWN_TIMEOUT` | `15s` | 优雅关闭等待 |
| `ENABLE_METRICS` | `true` | 指标开关 |
| `APP_PORT` | `8080` | 部署脚本用的端口 |

## 本地开发

```bash
make          # 格式化+检查+测试+编译
make test     # 单元测试
make test-race
make build    # 编译到 bin/server
make run      # 本地运行
make help     # 所有命令
```

## 部署

推送到 `main` 即自动部署。手动部署：

```bash
# 本地打包上传到机器
tar czf /tmp/app.tar.gz --exclude=.git .
cat /tmp/app.tar.gz | instabox exec wellknow-box --stdin -- \
  bash -c 'tar xz -C /data/instabox-go-starter'
# 机器执行部署
instabox exec wellknow-box -- bash -lc \
  'cd /data/instabox-go-starter && bash scripts/deploy.sh'
```

## 运维

```bash
# 在机器上
/data/gobin/run/server.pid        # PID 文件
/data/gobin/logs/server.log       # 业务日志
/data/gobin/logs/deploy.log       # 部署日志
/data/gobin/logs/alerts.log       # 告警日志
bash scripts/watchdog.sh          # 手动启动自愈守护
```

## 仓库结构

```
├── cmd/server/main.go            # 入口
├── internal/config/              # 配置 + 测试
├── internal/handler/             # 路由/处理器 + 测试
├── internal/server/              # 服务器 + 中间件
├── scripts/
│   ├── deploy.sh                 # 部署（含回滚）
│   └── watchdog.sh               # 自愈守护
├── .github/workflows/
│   ├── ci.yml
│   └── cd.yml
├── Makefile
├── Dockerfile
└── .golangci.yml
```
