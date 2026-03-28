# k3s VPS 安装步骤

本文档适用于单台 `Ubuntu/Debian` VPS，目标是部署当前仓库的 Go 服务，并配合 GitHub Actions 自动发布到 `k3s`。

执行时分成两部分：

- `本地执行`：在你自己的电脑上执行
- `服务器执行`：通过 `ssh dwj-vps` 登录 VPS 后执行

## 本地执行：准备部署密钥

### 1. 生成 GitHub Actions 部署密钥

作用：

- 生成一对专门给 GitHub Actions 使用的 SSH 密钥
- 私钥稍后放到 GitHub Secret `VPS_SSH_KEY`
- 公钥稍后安装到 VPS，允许 Actions 登录你的服务器

```bash
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/hello-go-actions
```

会生成两个文件：

- 私钥：`~/.ssh/hello-go-actions`
- 公钥：`~/.ssh/hello-go-actions.pub`

### 2. 把公钥安装到 VPS

作用：

- 把刚才生成的公钥追加到服务器的 `~/.ssh/authorized_keys`
- 这样持有对应私钥的人就可以 SSH 登录 `dwj-vps`

```bash
ssh-copy-id -i ~/.ssh/hello-go-actions.pub dwj-vps
```

### 3. 记录 GitHub Secrets

作用：

- 后面 GitHub Actions 需要靠这 4 个值 SSH 登录你的服务器
- 这些值会填到 GitHub 仓库的 `Settings -> Secrets and variables -> Actions`

后面要在 GitHub 仓库里配置这些值：

- `VPS_HOST`：VPS 公网 IP 或域名
- `VPS_PORT`：SSH 端口，通常为 `22`
- `VPS_USER`：`duwenjie`
- `VPS_SSH_KEY`：`~/.ssh/hello-go-actions` 私钥全文

## 服务器执行：安装 k3s 并初始化

### 1. 连接服务器

作用：

- 登录你的 VPS，后续所有“服务器执行”的命令都在这台机器上运行

```bash
ssh dwj-vps
```

### 2. 基础初始化

作用：

- 更新系统包索引
- 安装后面会用到的基础工具
- 设置时区，方便看日志和排障

```bash
sudo apt update
sudo apt install -y curl wget vim git ufw
sudo timedatectl set-timezone Asia/Shanghai
```

可选检查：

作用：

- `free -h`：查看内存是否够用
- `df -h`：查看磁盘空间
- `ss -lntp`：查看当前端口占用，避免和后面服务冲突

```bash
free -h
df -h
ss -lntp
```

### 3. 安装 k3s

当前方案使用单节点 `k3s`，关闭 `servicelb` 和 `traefik`，减少资源占用：

说明：

- `k3s` 是轻量 Kubernetes，适合低配 VPS 学习
- `--disable servicelb`：关闭内置负载均衡组件，当前用不到
- `--disable traefik`：关闭默认 Ingress，当前你先走 `IP + NodePort`

```bash
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--disable servicelb --disable traefik" sh -
```

验证：

说明：

- 第一条看节点是否 Ready
- 第二条看系统 Pod 是否正常运行

```bash
sudo k3s kubectl get nodes
sudo k3s kubectl get pods -A
```

### 4. 给当前用户配置 kubectl

当前使用的 SSH 用户是 `duwenjie`，直接给这个用户配置 `kubectl`：

作用：

- 把 `k3s` 的 kubeconfig 复制到当前用户目录
- 这样你后面不必每次都写 `sudo k3s kubectl`

```bash
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown "$(id -u):$(id -g)" ~/.kube/config
chmod 600 ~/.kube/config
```

如果系统中没有 `kubectl` 命令，补一个软链：

说明：

- `k3s` 自带 kubectl 能力
- 这条软链只是让你直接输入 `kubectl` 更方便

```bash
sudo ln -sf /usr/local/bin/k3s /usr/local/bin/kubectl
kubectl get nodes
```

### 5. 开放防火墙端口

当前仓库默认通过 `NodePort 30080` 暴露服务：

说明：

- `22/tcp`：给 SSH 用
- `30080/tcp`：给你的 Go 服务对外访问用

```bash
sudo ufw allow 22/tcp
sudo ufw allow 30080/tcp
sudo ufw enable
sudo ufw status
```

如果云厂商还有安全组，也要同步放行：

- `22/tcp`
- `30080/tcp`

### 6. 创建命名空间

作用：

- 给这个项目单独留一个 Kubernetes 逻辑空间
- 以后查看 Pod、Deployment、Service 都会更清晰

```bash
kubectl create namespace hello-go
```

如果已存在可以忽略错误。

### 7. 验证 SSH 密钥和 kubectl

这一步在本地执行，确认 GitHub Actions 后续能正常登录：

说明：

- 第一条验证刚才生成的私钥是否真能登录 VPS
- 第二条验证登录后 `kubectl` 是否可用

```bash
ssh -i ~/.ssh/hello-go-actions dwj-vps
kubectl get ns
```

## 本地执行：配置 GitHub

### 1. 配置 GitHub Secrets

在 GitHub 仓库 `Settings > Secrets and variables > Actions` 中添加：

- `VPS_HOST`：`39.97.61.187`
- `VPS_PORT`
- `VPS_USER`：`duwenjie`
- `VPS_SSH_KEY`

说明：

- `VPS_HOST`：你的服务器地址
- `VPS_PORT`：你的 SSH 端口，默认就是 `22`
- `VPS_USER`：GitHub Actions 登录服务器时使用的用户名
- `VPS_SSH_KEY`：本地文件 `~/.ssh/hello-go-actions` 的完整私钥内容

查看私钥内容：

```bash
cat ~/.ssh/hello-go-actions
```

复制时要把下面整段完整复制进去：

```text
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

### 2. 设置 GHCR 包为公开

当前部署配置默认使用公开 GHCR 镜像，不创建 `imagePullSecret`。

需要确认：

- 仓库 Actions 有权限推送 GHCR
- `ghcr.io/jaxgg/hello-go` 包可公开拉取

### 3. 首次部署验证

合并一次 `main` 后，GitHub Actions 会：

1. 在 `CI` 中执行 `go test ./...`
2. 构建并推送 `ghcr.io/jaxgg/hello-go`
3. 在 `CD` 中通过 SSH 登录 VPS
4. 执行 `kubectl apply` 和 `kubectl set image`

部署完成后，在 VPS 上验证：

说明：

- 第一条看 Deployment、Pod、Service 是否都起来了
- 第二条直接访问公网地址，验证服务是否真的对外可用

```bash
kubectl -n hello-go get deploy,pod,svc
curl http://<VPS_IP>:30080
```

预期返回：

```text
Hello World from Go!
```

## 服务器执行：常用排障命令

说明：

- `get pods -o wide`：看 Pod 跑在哪、IP 是什么
- `describe pod`：看启动失败原因、事件信息
- `logs deploy/hello`：看应用日志
- `rollout status`：看发布是否卡住
- `rollout undo`：回滚到上一个版本

```bash
kubectl -n hello-go get pods -o wide
kubectl -n hello-go describe pod <pod-name>
kubectl -n hello-go logs deploy/hello
kubectl -n hello-go rollout status deployment/hello --timeout=120s
kubectl -n hello-go rollout undo deployment/hello
```
