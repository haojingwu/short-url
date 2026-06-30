# ============================================================
# 第一阶段：构建阶段（builder）
# ============================================================
FROM golang:1.26-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统工具
RUN apk add --no-cache git ca-certificates

# 复制 go.mod 和 go.sum（利用 Docker 缓存）
COPY go.mod go.sum ./
#RUN go mod download
# 先查看哪些包需要下载
RUN go mod graph

# 尝试下载，并打印详细日志
RUN go mod download -x

# 复制源代码
COPY . .

# 编译二进制文件
# -ldflags="-s -w" 可以减小二进制体积
# -s: 去掉符号表
# -w: 去掉调试信息
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o short-url main.go

# ============================================================
# 第二阶段：运行阶段（runner）
# ============================================================
FROM alpine:latest

# 安装 ca-certificates（用于 HTTPS 请求）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为 Asia/Shanghai
ENV TZ=Asia/Shanghai

# 创建非 root 用户（安全最佳实践）
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/short-url .

# 从构建阶段复制配置文件（如果有）
# COPY --from=builder /app/config.yaml .

# 更改文件所有者
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 启动服务
CMD ["./short-url"]