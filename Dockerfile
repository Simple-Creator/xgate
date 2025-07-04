# -------- 前端构建阶段 --------
FROM node:18-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend .
RUN npm run build

# -------- 后端构建阶段 --------
FROM golang:alpine AS backend-builder
WORKDIR /backend
RUN apk add --no-cache gcc musl-dev
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o jump-backend .

# -------- 生产镜像 --------
FROM nginx:stable-alpine
WORKDIR /app

# 复制前端静态文件到 nginx
COPY --from=frontend-builder /frontend/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# 复制后端二进制和配置
COPY --from=backend-builder /backend/jump-backend /app/jump-backend

# 设置 Gin 为 release 模式
ENV GIN_MODE=release

# 暴露端口并启动服务
EXPOSE 8080
CMD ["/bin/sh", "-c", "/app/jump-backend & nginx -g 'daemon off;'"] 