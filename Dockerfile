# 多阶段构建 - 后端 Go 服务
FROM golang:1.23-alpine AS builder

WORKDIR /build

# 安装必要工具
RUN apk add --no-cache git

# 复制 go.mod (和可选的 go.sum)
COPY go.mod ./
COPY go.sum* ./
RUN go mod download || (go mod tidy && go mod download)

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server ./cmd/server

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装必要运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/server /app/server

# 创建日志和配置目录
RUN mkdir -p /app/logs /app/configs

# 暴露端口
EXPOSE 8080 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# 运行应用
CMD ["/app/server"]
