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
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o server ./cmd/server

# Stage 3: 最终运行镜像
FROM alpine:3.19
RUN apk add --no-cache nginx ca-certificates tzdata sqlite-libs \
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

# 创建启动脚本
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'nginx' >> /app/start.sh && \
    echo 'exec /app/server' >> /app/start.sh && \
    chmod +x /app/start.sh

EXPOSE 80

CMD ["/app/start.sh"]