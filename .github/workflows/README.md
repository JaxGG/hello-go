# CI/CD 配置说明

## GitHub Secrets 配置

在 GitHub 仓库的 Settings > Secrets and variables > Actions 中配置以下 secrets：

### 必需配置

1. **KUBECONFIG** (必需)
   - 你的 Kubernetes 集群的 kubeconfig 文件内容（base64 编码）
   - 获取方式：
     ```bash
     cat ~/.kube/config | base64
     ```

2. **KUBERNETES_NAMESPACE** (可选)
   - 部署的目标命名空间，默认为 `default`
   - 如果使用默认命名空间，可以不配置

### 镜像仓库

- 使用 GitHub Container Registry (ghcr.io)，自动使用 `GITHUB_TOKEN`
- 如需使用 Docker Hub，修改 workflow 中的 `REGISTRY` 和登录步骤

## 部署流程

1. **Push 到 main/master 分支**：自动构建镜像并部署到 K8s
2. **Pull Request**：只构建镜像，不部署
3. **手动触发**：可在 Actions 页面手动运行 workflow

## 更新镜像名称

在 `k8s/deployment.yaml` 中，将 `OWNER` 替换为你的 GitHub 用户名或组织名。

