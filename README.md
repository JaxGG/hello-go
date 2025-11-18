# Hello Go - Kubernetes CI/CD 示例

这是一个完整的 Go 应用 Kubernetes 部署示例，包含 GitHub Actions CI/CD 配置。

## 功能特性

- ✅ Go HTTP 服务器
- ✅ Docker 多阶段构建
- ✅ GitHub Actions CI/CD
- ✅ 自动构建和推送镜像到 GitHub Container Registry
- ✅ 自动部署到 Kubernetes

## 快速开始

### 本地运行

```bash
# 运行应用
CGO_ENABLED=0 go run main.go

# 或构建后运行
CGO_ENABLED=0 go build -o hello-go main.go
./hello-go
```

访问 http://localhost:8080

### Docker 构建

```bash
docker build -t hello-go:latest .
docker run -p 8080:8080 hello-go:latest
```

### Kubernetes 部署

#### 1. 配置 GitHub Secrets

在 GitHub 仓库 Settings > Secrets and variables > Actions 中添加：

- **KUBECONFIG**: Kubernetes 集群配置（base64 编码）
  ```bash
  cat ~/.kube/config | base64
  ```
- **KUBERNETES_NAMESPACE** (可选): 目标命名空间，默认为 `default`

#### 2. 更新镜像名称

编辑 `k8s/deployment.yaml`，将 `OWNER` 替换为你的 GitHub 用户名或组织名：

```yaml
image: ghcr.io/jaxgg/hello-go:latest
```

#### 3. 手动部署（可选）

```bash
kubectl apply -f k8s/deployment.yaml
kubectl get pods -l app=hello
kubectl get svc hello
```

#### 4. 访问服务

```bash
# 如果使用 NodePort
kubectl port-forward svc/hello 8080:8080
```

然后访问 http://localhost:8080

## CI/CD 流程

### 自动触发

- **Push 到 main/master**: 构建镜像并自动部署
- **Pull Request**: 仅构建镜像，不部署

### 手动触发

在 GitHub Actions 页面可以手动运行 workflow。

## 项目结构

```
.
├── .github/
│   └── workflows/
│       └── ci-cd.yml          # GitHub Actions CI/CD 配置
├── k8s/
│   └── deployment.yaml        # Kubernetes 部署配置
├── Dockerfile                 # Docker 镜像构建文件
├── main.go                    # Go 应用主文件
└── go.mod                     # Go 模块配置
```

## 镜像仓库

默认使用 GitHub Container Registry (ghcr.io)，如需使用 Docker Hub：

1. 修改 `.github/workflows/ci-cd.yml` 中的 `REGISTRY` 为 `docker.io`
2. 添加 `DOCKER_USERNAME` 和 `DOCKER_PASSWORD` secrets
3. 更新登录步骤使用 Docker Hub 凭据


