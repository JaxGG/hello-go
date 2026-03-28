# Kubernetes 配置说明

当前主部署路径面向单机 `k3s`：

- 使用 `namespace.yaml` 创建 `hello-go` 命名空间
- 使用 `deployment.yaml` 创建应用和 `NodePort` Service
- 通过 `GitHub Actions -> SSH -> kubectl` 自动部署

## 当前使用的文件

- `namespace.yaml`：创建 `hello-go` 命名空间
- `deployment.yaml`：创建 `Deployment` 和 `Service`

## 部署命令

首次初始化或手动重放时：

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/deployment.yaml
```

自动部署时，GitHub Actions 会在 VPS 上执行：

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/deployment.yaml
kubectl -n hello-go set image deployment/hello hello=ghcr.io/jaxgg/hello-go:sha-<short_sha>
kubectl -n hello-go rollout status deployment/hello --timeout=120s
```

## 访问方式

服务通过 `NodePort` 暴露：

```bash
curl http://<VPS_IP>:30080
```

## 说明

- `deployment.yaml` 中的默认镜像是 `ghcr.io/jaxgg/hello-go:latest`
- 实际自动部署时，CD 会覆盖为对应的 commit tag
- 仓库已经移除了旧的 `kind/webhook` 方案和当前不使用的可选清单，只保留主部署路径所需文件
