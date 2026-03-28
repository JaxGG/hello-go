# CI/CD 配置说明

当前流程面向单机 `k3s` 和 `GHCR`：

- `CI`：测试、构建镜像、推送到 `ghcr.io`
- `CD`：通过 SSH 登录 VPS，执行 `kubectl` 部署到 `k3s`

## GitHub Secrets

在仓库 `Settings > Secrets and variables > Actions` 中配置：

- `VPS_HOST`：VPS 公网 IP 或域名
- `VPS_PORT`：SSH 端口，通常是 `22`
- `VPS_USER`：部署用户，例如 `deploy`
- `VPS_SSH_KEY`：部署用户的私钥内容

当前默认使用公开 GHCR 镜像，因此不需要额外的镜像拉取凭据。

## Workflow 说明

### `ci.yml`

- `push main`：运行测试并推送镜像
- `pull_request -> main`：运行测试并构建镜像，但不推送

镜像标签规则：

- `latest`
- `sha-<short_sha>`

### `cd.yml`

- `CI` 在 `main` 分支成功完成后自动触发
- 先把仓库里的 `k8s/` 目录复制到 VPS
- 再通过 SSH 在 VPS 上执行：

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/deployment.yaml
kubectl -n hello-go set image deployment/hello hello=ghcr.io/jaxgg/hello-go:sha-<short_sha>
kubectl -n hello-go rollout status deployment/hello --timeout=120s
```

## 首次初始化

部署前先在 VPS 上完成：

```bash
kubectl create namespace hello-go
```

如果 namespace 已经存在，可以忽略报错；后续 CD 会使用 `kubectl apply` 保持幂等。

## 回滚

如果某次部署异常，可在 VPS 上执行：

```bash
kubectl -n hello-go rollout undo deployment/hello
kubectl -n hello-go rollout status deployment/hello --timeout=120s
```
