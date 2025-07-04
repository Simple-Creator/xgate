# ---- Build Stage ----
# 使用官方的 Go 镜像作为构建环境
FROM golang:alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装 cgo 依赖
RUN apk add --no-cache gcc musl-dev

# 复制所有项目文件
# 这一步必须在 go mod tidy 之前，以确保它能分析所有源文件
COPY . .

# 运行 go mod tidy 来下载依赖并确保 go.sum 文件是最新的
RUN go mod tidy

# 构建 Go 应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o jump-backend .

# ---- Release Stage ----
# 使用一个非常小的基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的应用二进制文件
COPY --from=builder /app/jump-backend .

# (可选) 如果您希望将配置文件打包到镜像中，可以在这里复制
# COPY --from=builder /app/config ./config

# 暴露应用监听的端口
EXPOSE 8080

# 容器启动时运行的命令
CMD ["./jump-backend"] 