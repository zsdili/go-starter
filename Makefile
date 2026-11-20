APP       := server
MODULE    := github.com/zsdili/go-starter
VERSION   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
COMMIT    ?= $(VERSION)
BUILD_AT  := $(shell date -u +%FT%TZ)
LDFLAGS   := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildAt=$(BUILD_AT)
BIN_DIR   := bin

.PHONY: all build run test test-race cover vet fmt lint clean docker deploy help

all: fmt vet test build ## 默认：格式化+检查+测试+编译

build: ## 编译当前平台二进制
	@mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP) ./cmd/server
	@echo "构建完成: $(BIN_DIR)/$(APP)"

run: ## 本地运行
	go run ./cmd/server

test: ## 运行单元测试
	go test ./...

test-race: ## 竞态检测
	go test -race ./...

cover: ## 生成覆盖率报告
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

vet: ## 静态分析
	go vet ./...

fmt: ## 格式化代码
	gofmt -w .

lint: ## 代码规范检查（需安装 golangci-lint）
	@command -v golangci-lint >/dev/null || { echo "请先安装 golangci-lint"; exit 1; }
	golangci-lint run ./...

clean: ## 清理构建产物
	rm -rf $(BIN_DIR) dist coverage.out coverage.html

docker: ## 构建 Docker 镜像
	docker build -t $(MODULE):$(VERSION) .

deploy: ## 在 InstaBox 机器上部署（需先上传源码）
	bash scripts/deploy.sh

help: ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n",$$1,$$2}'
