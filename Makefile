#Makefile - 项目构建和测试命令

.PHONY: help test test-cover build clean

# 默认目标: 显示帮助
help:
	@echo "可用命令:"
	@echo "  make test        运行单元测试"
	@echo "  make test-cover  运行测试并显示覆盖率"
	@echo "  make test-html   运行测试并生成 HTML 覆盖率报告"
	@echo "  make build       编译项目"
	@echo "  make run         运行项目"
	@echo "  make clean       清理编译产物"

# 运行单元测试
test:
	@echo "🧪 运行单元测试..."
	go test -v ./...

# 运行测试并显示覆盖率
test-cover:
	@echo "📊 运行测试并生成覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# 运行测试并生成 HTML 覆盖率报告
test-html:
	@echo "🌐 生成 HTML 覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 报告已生成: coverage.html"

# 编译项目
build:
	@echo "🔨 编译项目..."
	go build -o short-url main.go

# 运行项目
run:
	@echo "🚀 启动服务..."
	go run main.go

# 清理编译产物
clean:
	@echo "🧹 清理编译产物..."
	rm -f short-url
	rm -f coverage.out coverage.html