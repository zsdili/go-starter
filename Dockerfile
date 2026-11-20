# 多阶段构建：编译阶段用 golang 镜像，运行阶段用极小镜像
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_AT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildAt=${BUILD_AT}" \
    -o /server ./cmd/server

# 运行阶段：scratch 之上加 CA 证书和非 root 用户
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S app && adduser -S -G app app
COPY --from=builder /server /usr/local/bin/server
USER app
EXPOSE 8080
ENV APP_ADDR=:8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/server"]
