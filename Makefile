.PHONY: all build test clean fmt lint run help

# 获取当前 Git 标签作为版本号
VERSION := $(shell git describe --tags --always --dirty)
# 获取当前 Git 提交的哈希值
COMMIT := $(shell git rev-parse --short HEAD)
# 构建时间
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
# 主程序名称
BINARY_NAME := agent-forge
# Go 构建标签
BUILD_FLAGS := -ldflags "-X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

all: fmt lint test build

# 构建应用
build:
	@echo "构建应用..."
	@go build $(BUILD_FLAGS) -o $(BINARY_NAME)

# 运行测试
test:
	@echo "运行测试..."
	@go test -v -race -cover ./...

# 运行基准测试
bench:
	@echo "运行基准测试..."
	@go test -v -bench=. -benchmem ./...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -f $(BINARY_NAME)
	@go clean -testcache

# 格式化代码
fmt:
	@echo "格式化代码..."
	@go fmt ./...
	@gofmt -s -w .

# 运行代码检查
lint:
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "请先安装 golangci-lint"; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# 安装依赖
deps:
	@echo "安装依赖..."
	@go mod download
	@go mod tidy

# 运行应用
run:
	@echo "运行应用..."
	@go run $(BUILD_FLAGS) main.go

# 生成 mock 文件
mock:
	@echo "生成 mock 文件..."
	@if command -v mockgen >/dev/null 2>&1; then \
		go generate ./...; \
	else \
		echo "请先安装 mockgen"; \
		echo "go install github.com/golang/mock/mockgen@latest"; \
		exit 1; \
	fi

# 显示帮助信息
help:
	@echo "可用的 make 命令："
	@echo "  make build    - 构建应用"
	@echo "  make test     - 运行测试"
	@echo "  make bench    - 运行基准测试"
	@echo "  make clean    - 清理构建文件"
	@echo "  make fmt      - 格式化代码"
	@echo "  make lint     - 运行代码检查"
	@echo "  make deps     - 安装依赖"
	@echo "  make run      - 运行应用"
	@echo "  make mock     - 生成 mock 文件"
	@echo "  make help     - 显示帮助信息" 