# 全局构建参数
ARG VERSION=dev

# Stage 1: 构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci --registry=https://registry.npmmirror.com
COPY web/ ./
RUN npm run build

# Stage 2: 构建后端
FROM golang:1.24-alpine AS backend-builder
ARG TARGETARCH
ARG VERSION=dev
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w -X main.Version=${VERSION}" -o server ./cmd/server

# Stage 3: 构建 daemon 守护程序（多平台）
FROM golang:1.24-alpine AS daemon-builder
ARG VERSION=dev
WORKDIR /app
COPY frpc-daemon-ws/go.mod frpc-daemon-ws/go.sum ./
RUN go mod download
COPY frpc-daemon-ws/ ./
# 使用版本号构建多平台二进制文件
RUN mkdir -p /output && \
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.BuildTime=${VERSION}" -o /output/frpc-daemon-ws-linux-amd64 && \
    GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.BuildTime=${VERSION}" -o /output/frpc-daemon-ws-linux-arm64 && \
    GOOS=linux GOARCH=arm go build -ldflags="-s -w -X main.BuildTime=${VERSION}" -o /output/frpc-daemon-ws-linux-arm && \
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.BuildTime=${VERSION}" -o /output/frpc-daemon-ws-windows-amd64.exe && \
    GOOS=windows GOARCH=386 go build -ldflags="-s -w -X main.BuildTime=${VERSION}" -o /output/frpc-daemon-ws-windows-386.exe && \
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.BuildTime=${VERSION}" -o /output/frpc-daemon-ws-darwin-amd64 && \
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.BuildTime=${VERSION}" -o /output/frpc-daemon-ws-darwin-arm64

# Stage 4: 最终运行镜像
FROM alpine:3.19

# 设置默认时区
ENV TZ=Asia/Shanghai

ARG TARGETARCH
ARG S6_OVERLAY_VERSION=3.2.1.0

# 安装依赖并下载 s6-overlay（需要映射架构名称）
RUN apk add --no-cache nginx ca-certificates tzdata sqlite-libs xz curl \
    && S6_ARCH=$(case ${TARGETARCH} in \
    amd64) echo "x86_64" ;; \
    arm64) echo "aarch64" ;; \
    arm) echo "arm" ;; \
    *) echo ${TARGETARCH} ;; \
    esac) \
    && curl -fsSL "https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-noarch.tar.xz" -o /tmp/s6-overlay-noarch.tar.xz \
    && curl -fsSL "https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-${S6_ARCH}.tar.xz" -o /tmp/s6-overlay-arch.tar.xz \
    && tar -C / -Jxpf /tmp/s6-overlay-noarch.tar.xz \
    && tar -C / -Jxpf /tmp/s6-overlay-arch.tar.xz \
    && rm -f /tmp/s6-overlay-*.tar.xz \
    && mkdir -p /app/data /app/data/daemon /app/configs /var/run/nginx /usr/share/nginx/html

WORKDIR /app

# 复制后端二进制文件
COPY --from=backend-builder /app/server /app/server
COPY --from=backend-builder /app/configs/config.yaml /app/configs/config.yaml

# 复制前端静态文件
COPY --from=frontend-builder /app/web/dist /usr/share/nginx/html

# 复制 nginx 配置
COPY web/nginx.conf /etc/nginx/http.d/default.conf

# 复制数据文件
COPY backend/data/ip2region_v4.xdb /app/data/ip2region_v4.xdb

# 复制 daemon 守护程序二进制文件
COPY --from=daemon-builder /output/ /app/data/daemon/

# 复制 s6-overlay 服务定义
COPY docker/s6-rc.d /etc/s6-overlay/s6-rc.d

# 设置脚本执行权限
RUN chmod +x /etc/s6-overlay/s6-rc.d/nginx/run \
    && chmod +x /etc/s6-overlay/s6-rc.d/server/run

EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/api/health || exit 1

ENTRYPOINT ["/init"]