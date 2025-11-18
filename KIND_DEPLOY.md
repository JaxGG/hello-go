# 使用 Kind 在本地部署 Kubernetes 集群

本文档介绍如何在 macOS 上使用 Kind 部署 hello-go 项目。

## 前置要求

确保已安装以下工具：

```bash
# 检查 Docker
docker --version

# 检查 kubectl
kubectl version --client

# 检查 kind
kind version
```

## 安装 Kind

如果未安装，使用 Homebrew 安装：

```bash
brew install kind
```

## 部署步骤

### 1. 创建 Kind 集群

```bash
kind create cluster --name hello-go-cluster --config kind-config.yaml
```

### 2. 配置 GitHub Container Registry 认证（如果需要）

如果镜像仓库是私有的，需要配置认证：

```bash
# 创建 GitHub Personal Access Token (需要 package:read 权限)
# 访问: https://github.com/settings/tokens

# 创建 Secret
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN \
  --docker-email=YOUR_EMAIL

# 如果镜像仓库是公开的，可以跳过此步骤
```

### 3. 部署应用

```bash
kubectl apply -f k8s/deployment.yaml
```

### 4. 等待 Pod 就绪

```bash
kubectl wait --for=condition=ready pod -l app=hello --timeout=60s
```

### 5. 查看状态

```bash
kubectl get all
```

### 6. 测试访问

```bash
# 方法 1：使用端口转发
kubectl port-forward svc/hello 8080:8080

# 在另一个终端测试
curl http://localhost:8080

# 方法 2：通过 NodePort
curl http://localhost:30080
```

## 常用命令

### 查看资源

```bash
# 查看所有资源
kubectl get all

# 查看 Pod 日志
kubectl logs -l app=hello

# 查看 Pod 详细信息
kubectl describe pod -l app=hello
```

### 更新部署

```bash
# 当 GitHub Action 推送新镜像后，重启部署拉取最新镜像
kubectl rollout restart deployment/hello

# 查看更新状态
kubectl rollout status deployment/hello

# 查看当前使用的镜像
kubectl describe deployment hello | grep Image
```

### 扩缩容

```bash
# 扩展副本数
kubectl scale deployment hello --replicas=3

# 查看副本状态
kubectl get deployment hello
```

## 清理

### 删除应用

```bash
kubectl delete -f k8s/deployment.yaml
```

### 删除集群

```bash
kind delete cluster --name hello-go-cluster
```

## 一键部署脚本

```bash
#!/bin/bash
set -e

echo "1. 创建集群..."
kind create cluster --name hello-go-cluster --config kind-config.yaml

echo "2. 部署应用..."
kubectl apply -f k8s/deployment.yaml

echo "3. 等待 Pod 就绪..."
kubectl wait --for=condition=ready pod -l app=hello --timeout=60s

echo "4. 查看状态..."
kubectl get all

echo "部署完成！"
echo "访问服务: kubectl port-forward svc/hello 8080:8080"
```

## 故障排查

### Pod 一直处于 Pending 状态

```bash
kubectl describe pod -l app=hello
```

### 镜像拉取失败

```bash
# 查看 Pod 事件
kubectl describe pod -l app=hello

# 如果是私有仓库，检查 Secret
kubectl get secret ghcr-secret

# 如果 Secret 不存在，需要创建（见步骤 2）
```

### 服务无法访问

```bash
# 检查 Service
kubectl get svc hello

# 检查 Endpoints
kubectl get endpoints hello

# 使用端口转发测试
kubectl port-forward svc/hello 8080:8080
```

## GitHub Action 集成

### 配置说明

项目已配置 GitHub Actions 工作流：
- **自动构建**：`ci-cd.yml` 在推送代码到 `main` 分支时自动构建镜像并推送到 GHCR
- **手动部署**：`deploy.yml` 通过 webhook 触发本地 kind 集群部署

### 1. 配置 GitHub Secrets

在 GitHub 仓库设置中添加以下 Secret：

- `WEBHOOK_URL`: 你的内网穿透 webhook 地址（例如：`http://your-tunnel-url/webhook`）

### 2. 配置 Docker 认证（如果镜像仓库是私有的）

如果 GHCR 镜像仓库是私有的，需要先登录：

```bash
# 创建 GitHub Personal Access Token (需要 package:read 权限)
# 访问: https://github.com/settings/tokens

# 登录到 GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin
```

### 3. 启动本地 Webhook 服务器

Webhook 服务器会自动从 GHCR 拉取镜像并加载到 kind 集群：

```bash
# 编译 webhook 服务器（使用静态链接，避免 macOS 动态链接问题）
cd scripts
CGO_ENABLED=0 go build -ldflags="-s -w" -o webhook-server webhook-server.go

# 启动服务器（默认端口 9000）
# 可以从任何目录运行，会自动查找项目根目录
./webhook-server

# 或指定端口、密钥和 kind 集群名称
PORT=9000 WEBHOOK_SECRET=your-secret KIND_CLUSTER_NAME=hello-go-cluster ./webhook-server
```

### 4. 配置内网穿透

使用你的内网穿透方案，将本地 webhook 服务器（端口 9000）暴露到公网。

然后将生成的 URL 配置到 GitHub Secrets 的 `WEBHOOK_URL`。

### 5. 工作流程

1. **自动构建**：推送代码到 `main` 分支，`ci-cd.yml` 自动触发，构建 Docker 镜像并推送到 GHCR
2. **自动部署**：`build-and-push` job 成功后，`deploy.yml` 自动触发（也可手动触发）
3. **Webhook 处理**：
   - GitHub Action 调用 webhook URL（通过内网穿透）
   - 本地 webhook 服务器接收请求
   - 从 GHCR 拉取镜像到本地 Docker
   - 使用 `kind load docker-image` 将镜像加载到 kind 集群
   - 更新 Kubernetes 部署使用本地镜像名（`hello-go:latest`）

**完整流程**：推送代码 → 自动构建 → 自动部署 → 本地集群更新

### 6. 手动触发部署

如果 webhook 失败，可以手动部署：

```bash
# 从 GHCR 拉取镜像
docker pull ghcr.io/jaxgg/hello-go:latest

# 打本地标签
docker tag ghcr.io/jaxgg/hello-go:latest hello-go:latest

# 加载到 kind 集群
kind load docker-image hello-go:latest --name hello-go-cluster

# 更新部署
kubectl set image deployment/hello hello=hello-go:latest

# 查看部署状态
kubectl rollout status deployment/hello
```

### 7. 测试 Webhook

可以手动测试 webhook 是否正常工作：

```bash
curl -X POST http://localhost:9000/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "image": "ghcr.io/jaxgg/hello-go:latest",
    "tag": "latest",
    "ref": "refs/heads/main",
    "commit": "abc123"
  }'
```
