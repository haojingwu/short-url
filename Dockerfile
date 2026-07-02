# ============================================================
# 第一阶段：构建阶段（builder）
# ============================================================
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o short-url main.go

# ============================================================
# 第二阶段：运行阶段（runner）
# ============================================================
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Shanghai

RUN adduser -D -g '' appuser

WORKDIR /app

# 复制二进制文件
COPY --from=builder /app/short-url .

# 复制配置文件
COPY --from=builder /app/config ./config

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./short-url"]