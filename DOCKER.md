# Docker 部署指南

## 前提条件

- 安装 [Docker](https://docs.docker.com/get-docker/)
- 安装 [Docker Compose](https://docs.docker.com/compose/install/) (可选，用于本地开发)
- 获取 DeepSeek API 密钥

## 使用 Docker 构建和运行

### 构建镜像

```bash
docker build -t agent-forge .
```

### 运行容器

```bash
docker run -d --name agent-forge \
  -p 8080:8080 \
  --env DEEPSEEK_API_KEY=your_api_key_here \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml \
  -i \
  agent-forge
```

> **重要:** `-i` 参数是必须的，它保持容器的标准输入通道打开，使MCP程序能够正常工作。如果没有此参数，容器可能会在启动后立即退出。

## 使用 Docker Compose (推荐)

### 准备环境变量

复制示例环境变量文件并填入你的 API 密钥:

```bash
cp .env.example .env
# 编辑 .env 文件，填入实际的 API 密钥
```

### 启动服务

```bash
docker-compose up -d
```

### 查看日志

```bash
docker-compose logs -f
```

### 停止服务

```bash
docker-compose down
```

## 配置说明

容器内配置文件位置: `/app/config/config.yaml`

你可以通过挂载自己的配置文件来覆盖默认配置:

```bash
docker run -d --name agent-forge \
  -p 8080:8080 \
  --env DEEPSEEK_API_KEY=your_api_key_here \
  -v $(pwd)/your-config.yaml:/app/config/config.yaml \
  -v $(pwd)/logs:/app/logs \
  -i \
  agent-forge
```

## 健康检查

容器内置了健康检查，可通过以下命令查看服务健康状态:

```bash
docker inspect --format='{{json .State.Health}}' agent-forge | jq
```

## 故障排除

1. 如果遇到权限问题，请检查挂载的卷权限
2. 确保 DeepSeek API 密钥设置正确
3. 检查服务日志:
   ```bash
   docker logs agent-forge
   ```

### 容器退出问题

如果容器在启动后退出：

1. **检查是否使用了 `-i` 参数**:
   ```bash
   # 正确的启动方式，带 -i 参数
   docker run -i -d --name agent-forge -p 8080:8080 -e DEEPSEEK_API_KEY=your_api_key agent-forge
   ```

2. **检查退出原因**：
   ```bash
   docker logs agent-forge
   ```

3. **进入运行中的容器排查**：
   ```bash
   docker exec -it agent-forge /bin/sh
   ```

4. **检查应用是否有执行权限**：
   ```bash
   docker exec -it agent-forge /bin/sh -c "ls -la /app/agent-forge"
   ```