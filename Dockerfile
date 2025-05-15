# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache git

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o agent-forge .

# 运行阶段
FROM alpine:3.18

# 添加CA证书，用于HTTPS请求
RUN apk add --no-cache ca-certificates tzdata procps && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建非root用户
RUN adduser -D -g '' appuser

# 创建应用所需的目录结构
RUN mkdir -p /app/logs /app/config
WORKDIR /app

# 从构建阶段复制编译好的应用
COPY --from=builder /app/agent-forge .

# 复制启动脚本和配置文件
COPY start.sh /app/
COPY internal/config/config.yaml /app/config/

# 设置适当的权限
RUN chmod +x /app/agent-forge /app/start.sh && chown -R appuser:appuser /app
USER appuser

# 设置环境变量
ENV AGENT_FORGE_ENV=production

# 声明卷，用于持久化日志
VOLUME ["/app/logs"]

# 暴露应用端口
EXPOSE 8080

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# 使用启动脚本
ENTRYPOINT ["/app/start.sh"]

# 指定DeepSeek API密钥需要通过环境变量传入
# 启动容器示例: docker run -i -d --name agent-forge -p 8080:8080 -e DEEPSEEK_API_KEY=your_api_key agent-forge 