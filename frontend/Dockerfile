# ---- Build Stage ----
# 使用 Node.js 镜像进行构建
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 package.json 和 package-lock.json (或 yarn.lock)
COPY package*.json ./

# 安装依赖
RUN npm install

# 复制所有源代码
COPY . .

# 构建应用
RUN npm run build

# ---- Release Stage ----
# 使用 nginx 镜像来托管静态文件
FROM nginx:stable-alpine

# 从构建阶段复制构建好的静态文件到 nginx 的默认托管目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置文件
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 80

# 启动 nginx
CMD ["nginx", "-g", "daemon off;"] 