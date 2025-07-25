# ---- Frontend Build Stage ----
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端依赖并安装
COPY frontend/package*.json ./
RUN npm install

# 复制前端代码并构建
COPY frontend/ ./
RUN npm run build

# ---- Backend Build Stage ----
FROM golang:alpine AS backend-builder

WORKDIR /app/backend

# 安装构建工具
RUN apk add --no-cache gcc musl-dev

# 复制后端代码并构建
COPY backend/ ./
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o xgate-backend .

# ---- Final Stage ----
FROM nginx:alpine

# 设置工作目录
WORKDIR /app

# 从前端构建阶段复制构建好的静态文件到 nginx 目录
COPY --from=frontend-builder /app/frontend/dist /usr/share/nginx/html

# 复制专为一体式镜像设计的 nginx 配置
COPY frontend/nginx-allinone.conf /etc/nginx/conf.d/default.conf

# 从后端构建阶段复制可执行文件
COPY --from=backend-builder /app/backend/xgate-backend /app/

# 创建数据目录
RUN mkdir -p /app/data

# 设置默认环境变量 - 默认使用 sqlite
ENV XGATE_DATABASE_TYPE=sqlite
ENV XGATE_SERVER_PORT=8080
ENV XGATE_JWT_SECRET=a_very_secret_key_for_all_in_one

# 创建启动脚本
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo '# 启动后端服务' >> /app/start.sh && \
    echo 'cd /app && nohup ./xgate-backend &' >> /app/start.sh && \
    echo '# 等待几秒钟以确保后端启动' >> /app/start.sh && \
    echo 'sleep 3' >> /app/start.sh && \
    echo '# 启动 nginx 前端服务' >> /app/start.sh && \
    echo 'nginx -g "daemon off;"' >> /app/start.sh && \
    chmod +x /app/start.sh

# 暴露端口
EXPOSE 80

# 启动服务
CMD ["/app/start.sh"] 