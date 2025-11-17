# Kubernetes 部署指南

本文档包含在服务器上部署 hello-go 到 minikube 的完整步骤，所有命令需要手动执行。

## 前置要求

- Linux 服务器（Ubuntu/Debian）
- 已安装 Docker
- 已安装 minikube
- 需要安装 kubectl
- GitHub 账号（用于推送镜像到 GitHub Container Registry）

---

## 第一步：安装 kubectl

### Ubuntu/Debian 系统

```bash
# 1. 下载最新版本的 kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# 2. 验证下载的文件
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl.sha256"

# 3. 验证校验和
echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check

# 4. 安装 kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# 5. 验证安装
kubectl version --client
```

### 如果上述方法失败，使用包管理器安装

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y kubectl

# 验证安装
kubectl version --client
```

---

## 第二步：启动 minikube

### 2.1 配置 Docker 镜像加速器（重要：解决镜像拉取失败）

如果遇到镜像拉取失败（如 `registry.k8s.io` 或 `gcr.io` 无法访问），需要配置国内镜像源：

#### 方法 1：配置 Docker 镜像加速器（推荐）

```bash
# 创建或编辑 Docker daemon 配置文件
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com"
  ]
}
EOF

# 重启 Docker 服务
sudo systemctl daemon-reload
sudo systemctl restart docker

# 验证配置
docker info | grep -A 10 "Registry Mirrors"
```

#### 方法 2：配置 Kubernetes 镜像代理

```bash
# 创建镜像代理配置
sudo mkdir -p /etc/containerd
sudo tee /etc/containerd/config.toml <<-'EOF'
[plugins."io.containerd.grpc.v1.cri".registry]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry.k8s.io"]
      endpoint = ["https://registry.aliyuncs.com/google_containers"]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."gcr.io"]
      endpoint = ["https://gcr.mirrors.ustc.edu.cn"]
EOF

# 如果使用 containerd，重启服务
sudo systemctl restart containerd
```

#### 方法 3：使用阿里云镜像仓库（临时方案）

如果上述方法不行，可以手动拉取镜像并重命名：

```bash
# 配置镜像映射（在启动 minikube 前执行）
# 这些命令会在启动时自动处理镜像拉取
```

### 2.2 检查 minikube 状态

```bash
# 检查 minikube 状态
minikube status
```

### 2.3 启动 minikube（内存不足时使用）

如果遇到内存不足错误（如 `RSRC_INSUFFICIENT_CONTAINER_MEMORY`），按以下步骤操作：

#### 方法 1：强制启动（推荐，适用于 Docker 有 1613MB 可用的情况）

```bash
# 先删除现有的 minikube 实例（如果有）
minikube delete

# 使用 --force 参数强制启动，忽略内存检查
minikube start --memory=1500mb --cpus=2 --force

# 如果仍然失败，尝试更少的内存
minikube start --memory=1400mb --cpus=1 --force
```

#### 方法 2：增加 Docker 内存限制

如果方法 1 失败，可以增加 Docker 的内存限制：

```bash
# 查看当前 Docker 内存限制
docker system info | grep -i memory

# 如果使用 Docker Desktop，在设置中增加内存分配
# 如果使用 Docker Engine，需要修改 Docker daemon 配置
```

#### 方法 3：使用更激进的配置

```bash
# 删除现有实例
minikube delete

# 使用最小配置强制启动
minikube start \
  --memory=1400mb \
  --cpus=1 \
  --driver=docker \
  --force \
  --extra-config=kubelet.max-pods=10
```

#### 方法 4：如果以上都失败，使用 none 驱动（需要 root 权限）

```bash
# 注意：none 驱动需要 root 权限，且直接在主机上运行
sudo minikube start --driver=none
```

### 2.4 正常启动（如果内存充足）

```bash
# 如果内存充足（>= 1800MB），可以直接启动
minikube start
```

### 2.5 验证集群状态

```bash
# 等待启动完成，验证集群状态
kubectl cluster-info
kubectl get nodes

# 查看 minikube 配置
minikube config view
```

### 2.6 如果启动失败，删除并重新创建

```bash
# 删除现有的 minikube 集群
minikube delete

# 清理 minikube 配置
minikube config unset memory
minikube config unset cpus

# 使用强制模式重新创建（忽略内存检查）
minikube start --memory=1500mb --cpus=2 --force

# 验证启动
minikube status
kubectl get nodes
```

### 2.7 故障排查：镜像拉取失败问题

如果遇到镜像拉取失败（如 `registry.k8s.io` 或 `gcr.io` 无法访问）：

#### 2.7.1 解决 kicbase 镜像拉取失败（gcr.io/k8s-minikube/kicbase）

如果遇到 `Unable to find image 'gcr.io/k8s-minikube/kicbase'` 错误：

**方法 1：使用镜像代理拉取 kicbase（推荐）**

```bash
# 1. 先配置 Docker 镜像加速器（见 2.1 节）
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com"
  ]
}
EOF
sudo systemctl daemon-reload
sudo systemctl restart docker

# 2. 尝试从其他源拉取 kicbase（如果镜像加速器支持 gcr.io）
# 或者使用代理环境变量
export HTTP_PROXY=http://your-proxy:port
export HTTPS_PROXY=http://your-proxy:port
docker pull gcr.io/k8s-minikube/kicbase:v0.0.48

# 3. 如果仍然失败，尝试使用 minikube 的 --base-image 参数
minikube delete
minikube start --memory=1500mb --cpus=2 --force --base-image="gcr.io/k8s-minikube/kicbase:v0.0.48"
```

**方法 2：手动下载并导入 kicbase 镜像**

```bash
# 1. 从 GitHub Releases 下载 kicbase tarball（minikube 会自动尝试，但可以手动下载）
# 访问：https://github.com/kubernetes/minikube/releases
# 下载对应版本的 kicbase tarball

# 2. 或者使用代理下载
wget https://github.com/kubernetes/minikube/releases/download/v1.34.0/kicbase.tar

# 3. 导入镜像
docker load < kicbase.tar

# 4. 重新启动 minikube
minikube delete
minikube start --memory=1500mb --cpus=2 --force
```

**方法 3：配置 Docker 代理（如果有 HTTP/HTTPS 代理）**

```bash
# 创建 Docker 代理配置
sudo mkdir -p /etc/systemd/system/docker.service.d
sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf <<-'EOF'
[Service]
Environment="HTTP_PROXY=http://your-proxy:port"
Environment="HTTPS_PROXY=http://your-proxy:port"
Environment="NO_PROXY=localhost,127.0.0.1"
EOF

# 重启 Docker
sudo systemctl daemon-reload
sudo systemctl restart docker

# 验证代理
docker info | grep -i proxy

# 然后重新启动 minikube
minikube delete
minikube start --memory=1500mb --cpus=2 --force
```

#### 2.7.2 解决 Kubernetes 组件镜像拉取失败

```bash
# 1. 检查网络连接
ping registry.k8s.io
ping gcr.io

# 2. 如果无法访问，配置镜像加速器（见 2.1 节）

# 3. 手动拉取镜像（使用阿里云镜像）
docker pull registry.aliyuncs.com/google_containers/kube-proxy:v1.34.0
docker pull registry.aliyuncs.com/google_containers/kube-scheduler:v1.34.0
docker pull registry.aliyuncs.com/google_containers/kube-controller-manager:v1.34.0
docker pull registry.aliyuncs.com/google_containers/etcd:3.6.4-0
docker pull registry.aliyuncs.com/google_containers/pause:3.10.1

# 4. 重命名镜像标签
docker tag registry.aliyuncs.com/google_containers/kube-proxy:v1.34.0 registry.k8s.io/kube-proxy:v1.34.0
docker tag registry.aliyuncs.com/google_containers/kube-scheduler:v1.34.0 registry.k8s.io/kube-scheduler:v1.34.0
docker tag registry.aliyuncs.com/google_containers/kube-controller-manager:v1.34.0 registry.k8s.io/kube-controller-manager:v1.34.0
docker tag registry.aliyuncs.com/google_containers/etcd:3.6.4-0 registry.k8s.io/etcd:3.6.4-0
docker tag registry.aliyuncs.com/google_containers/pause:3.10.1 registry.k8s.io/pause:3.10.1

# 5. 拉取 storage-provisioner 镜像
docker pull registry.aliyuncs.com/google_containers/storage-provisioner:v5
docker tag registry.aliyuncs.com/google_containers/storage-provisioner:v5 gcr.io/k8s-minikube/storage-provisioner:v5

# 6. 然后重新启动 minikube
minikube delete
minikube start --memory=1500mb --cpus=2 --force
```

或者使用一键脚本：

```bash
# 创建镜像拉取脚本
cat > pull-k8s-images.sh << 'EOF'
#!/bin/bash
images=(
  "registry.aliyuncs.com/google_containers/kube-proxy:v1.34.0"
  "registry.aliyuncs.com/google_containers/kube-scheduler:v1.34.0"
  "registry.aliyuncs.com/google_containers/kube-controller-manager:v1.34.0"
  "registry.aliyuncs.com/google_containers/etcd:3.6.4-0"
  "registry.aliyuncs.com/google_containers/pause:3.10.1"
  "registry.aliyuncs.com/google_containers/storage-provisioner:v5"
)

for image in "${images[@]}"; do
  docker pull $image
done

# 重命名标签
docker tag registry.aliyuncs.com/google_containers/kube-proxy:v1.34.0 registry.k8s.io/kube-proxy:v1.34.0
docker tag registry.aliyuncs.com/google_containers/kube-scheduler:v1.34.0 registry.k8s.io/kube-scheduler:v1.34.0
docker tag registry.aliyuncs.com/google_containers/kube-controller-manager:v1.34.0 registry.k8s.io/kube-controller-manager:v1.34.0
docker tag registry.aliyuncs.com/google_containers/etcd:3.6.4-0 registry.k8s.io/etcd:3.6.4-0
docker tag registry.aliyuncs.com/google_containers/pause:3.10.1 registry.k8s.io/pause:3.10.1
docker tag registry.aliyuncs.com/google_containers/storage-provisioner:v5 gcr.io/k8s-minikube/storage-provisioner:v5
EOF

chmod +x pull-k8s-images.sh
./pull-k8s-images.sh
```

### 2.8 故障排查：内存不足问题

如果仍然遇到 `RSRC_INSUFFICIENT_CONTAINER_MEMORY` 错误：

```bash
# 1. 检查 Docker 可用内存
docker system df
docker info | grep -i memory

# 2. 清理 Docker 资源（释放内存）
docker system prune -a --volumes

# 3. 检查是否有其他容器占用内存
docker ps -a
docker stats --no-stream

# 4. 停止不必要的容器
docker stop $(docker ps -q)

# 5. 然后重新尝试启动 minikube
minikube delete
minikube start --memory=1500mb --cpus=2 --force
```

---

## 第三步：配置 GitHub Container Registry

### 3.1 登录 GitHub Container Registry

```bash
# 使用 GitHub Personal Access Token 登录
# 需要先创建 Token: https://github.com/settings/tokens
# Token 需要 package:write 权限
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin

# 或者交互式登录
docker login ghcr.io
# 用户名：YOUR_GITHUB_USERNAME
# 密码：YOUR_GITHUB_TOKEN
```

### 3.2 修改 deployment.yaml 中的镜像地址

编辑 `k8s/deployment.yaml`，将 `OWNER` 替换为你的 GitHub 用户名或组织名：

```bash
# 使用 sed 替换（将 YOUR_GITHUB_USERNAME 替换为实际用户名）
sed -i 's/ghcr.io\/OWNER\/hello-go:latest/ghcr.io\/YOUR_GITHUB_USERNAME\/hello-go:latest/g' k8s/deployment.yaml

# 或者手动编辑文件
nano k8s/deployment.yaml
# 找到 image: ghcr.io/OWNER/hello-go:latest
# 替换为 image: ghcr.io/YOUR_GITHUB_USERNAME/hello-go:latest
```

---

## 第四步：构建 Docker 镜像

```bash
# 进入项目目录
cd /path/to/hello-go

# 构建镜像（替换 YOUR_GITHUB_USERNAME 为实际用户名）
docker build -t ghcr.io/YOUR_GITHUB_USERNAME/hello-go:latest .

# 验证镜像
docker images | grep hello-go
```

---

## 第五步：推送镜像到 GitHub Container Registry

```bash
# 推送镜像（替换 YOUR_GITHUB_USERNAME 为实际用户名）
docker push ghcr.io/YOUR_GITHUB_USERNAME/hello-go:latest
```

**注意**：如果推送失败，可能需要：
1. 在 GitHub 仓库设置中启用 Container Registry
2. 确保 Token 有 `write:packages` 权限
3. 如果是私有仓库，确保镜像设置为公开或配置访问权限

---

## 第六步：配置 minikube 拉取私有镜像（如果需要）

如果镜像设置为私有，需要配置 minikube 的镜像拉取密钥：

```bash
# 创建镜像拉取密钥
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN \
  --docker-email=YOUR_EMAIL

# 修改 deployment.yaml，添加 imagePullSecrets
# 在 spec.template.spec 下添加：
# imagePullSecrets:
# - name: ghcr-secret
```

如果镜像设置为公开，可以跳过此步骤。

---

## 第七步：部署到 Kubernetes

```bash
# 部署应用
kubectl apply -f k8s/deployment.yaml

# 查看部署状态
kubectl get pods -l app=hello

# 查看 Service
kubectl get svc hello

# 查看详细信息
kubectl get deployment hello
```

---

## 第八步：等待 Pod 就绪

```bash
# 等待 Pod 就绪（最多等待 60 秒）
kubectl wait --for=condition=ready pod -l app=hello --timeout=60s

# 如果超时，查看 Pod 状态
kubectl get pods -l app=hello

# 查看 Pod 详细信息（如果有问题）
kubectl describe pod -l app=hello

# 查看 Pod 日志
kubectl logs -l app=hello
```

---

## 第九步：访问应用

### 方式 1: NodePort（推荐）

```bash
# 获取 minikube IP
minikube ip

# 访问应用（替换 <MINIKUBE_IP> 为实际 IP）
curl http://<MINIKUBE_IP>:30080

# 例如：如果 minikube IP 是 192.168.49.2
curl http://192.168.49.2:30080
```

### 方式 2: Port Forward

```bash
# 端口转发
kubectl port-forward svc/hello 8080:8080

# 在另一个终端访问
curl http://localhost:8080
```

### 方式 3: minikube service（自动打开浏览器）

```bash
minikube service hello
```

---

## 常用管理命令

### 查看资源状态

```bash
# 查看 Pod
kubectl get pods -l app=hello

# 查看 Service
kubectl get svc hello

# 查看 Deployment
kubectl get deployment hello

# 查看所有资源
kubectl get all -l app=hello
```

### 查看日志

```bash
# 查看 Pod 日志
kubectl logs -l app=hello

# 实时查看日志
kubectl logs -f -l app=hello

# 查看特定 Pod 的日志
kubectl logs <POD_NAME>
```

### 扩缩容

```bash
# 扩容到 3 个副本
kubectl scale deployment hello --replicas=3

# 查看副本状态
kubectl get pods -l app=hello
```

### 更新应用

```bash
# 1. 修改代码后重新构建镜像
docker build -t ghcr.io/YOUR_GITHUB_USERNAME/hello-go:latest .

# 2. 推送新镜像
docker push ghcr.io/YOUR_GITHUB_USERNAME/hello-go:latest

# 3. 重启 Deployment（拉取新镜像）
kubectl rollout restart deployment hello

# 4. 查看更新状态
kubectl rollout status deployment hello
```

### 卸载应用

```bash
# 删除部署
kubectl delete -f k8s/deployment.yaml

# 或者删除所有相关资源
kubectl delete deployment hello
kubectl delete svc hello
```

---

## 故障排查

### Pod 无法启动

```bash
# 查看 Pod 状态
kubectl get pods -l app=hello

# 查看 Pod 详细信息
kubectl describe pod <POD_NAME>

# 查看 Pod 日志
kubectl logs <POD_NAME>

# 常见问题：
# - ImagePullBackOff: 镜像拉取失败，检查镜像地址和权限
# - CrashLoopBackOff: 应用崩溃，查看日志
# - Pending: 资源不足或调度问题
```

### 镜像拉取失败

```bash
# 检查镜像是否存在
docker pull ghcr.io/YOUR_GITHUB_USERNAME/hello-go:latest

# 检查镜像拉取密钥
kubectl get secret ghcr-secret

# 如果镜像为私有，确保已创建 imagePullSecrets
kubectl describe pod <POD_NAME> | grep -A 5 "Image Pull Secrets"
```

### 无法访问服务

```bash
# 1. 检查 Service 是否正常
kubectl get svc hello
kubectl describe svc hello

# 2. 检查 Pod 是否运行
kubectl get pods -l app=hello

# 3. 检查端口映射
kubectl get svc hello -o yaml | grep nodePort

# 4. 测试 Pod 内部访问
kubectl exec -it <POD_NAME> -- wget -qO- http://localhost:8080
```

### 查看事件

```bash
# 查看所有事件
kubectl get events --sort-by='.lastTimestamp'

# 查看特定命名空间的事件
kubectl get events -n default
```

---

## 可选：安装 Ingress 和 HPA

### 安装 Ingress Controller

```bash
# 启用 minikube ingress 插件
minikube addons enable ingress

# 验证安装
kubectl get pods -n ingress-nginx
```

### 安装 metrics-server（HPA 需要）

```bash
# 启用 minikube metrics-server 插件
minikube addons enable metrics-server

# 验证安装
kubectl get pods -n kube-system | grep metrics-server
```

### 部署 Ingress 和 HPA

```bash
# 修改 ingress.yaml 中的域名（可选）
nano k8s/ingress.yaml

# 部署 Ingress
kubectl apply -f k8s/ingress.yaml

# 部署 HPA
kubectl apply -f k8s/hpa.yaml

# 查看状态
kubectl get ingress
kubectl get hpa
```

---

## 注意事项

1. **镜像地址**：确保 `k8s/deployment.yaml` 中的镜像地址正确，替换 `YOUR_GITHUB_USERNAME` 为实际用户名。

2. **GitHub Token**：推送镜像需要 GitHub Personal Access Token，确保有 `write:packages` 权限。

3. **镜像可见性**：如果镜像设置为私有，需要配置 `imagePullSecrets`。

4. **资源限制**：默认配置了 CPU 和内存限制，可根据服务器配置调整。

5. **健康检查**：已配置 liveness 和 readiness 探针，确保应用正常运行。

6. **防火墙**：如果使用 NodePort，确保服务器防火墙允许 30080 端口访问。

7. **内存配置**：如果 Docker 可用内存不足 1800MB，启动 minikube 时使用 `--memory=1500mb --force` 参数。

8. **镜像源问题**：如果遇到 `registry.k8s.io` 或 `gcr.io` 镜像拉取失败，需要配置 Docker 镜像加速器或手动拉取镜像（见 2.1 和 2.7 节）。

---

## 完整命令清单（快速参考）

```bash
# 1. 安装 kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# 2. 配置 Docker 镜像加速器（解决镜像拉取失败）
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com"
  ]
}
EOF
sudo systemctl daemon-reload
sudo systemctl restart docker

# 3. 如果镜像加速器无效，手动拉取 Kubernetes 镜像（可选）
# 注意：根据 minikube 实际使用的版本调整镜像版本号
# 先尝试启动 minikube，查看报错信息中的具体版本号，然后替换下面的版本
# 示例：如果报错显示 v1.34.0，则使用该版本
docker pull registry.aliyuncs.com/google_containers/kube-proxy:v1.34.0
docker pull registry.aliyuncs.com/google_containers/kube-scheduler:v1.34.0
docker pull registry.aliyuncs.com/google_containers/kube-controller-manager:v1.34.0
docker pull registry.aliyuncs.com/google_containers/etcd:3.6.4-0
docker pull registry.aliyuncs.com/google_containers/pause:3.10.1
docker pull registry.aliyuncs.com/google_containers/storage-provisioner:v5
# 重命名镜像标签
docker tag registry.aliyuncs.com/google_containers/kube-proxy:v1.34.0 registry.k8s.io/kube-proxy:v1.34.0
docker tag registry.aliyuncs.com/google_containers/kube-scheduler:v1.34.0 registry.k8s.io/kube-scheduler:v1.34.0
docker tag registry.aliyuncs.com/google_containers/kube-controller-manager:v1.34.0 registry.k8s.io/kube-controller-manager:v1.34.0
docker tag registry.aliyuncs.com/google_containers/etcd:3.6.4-0 registry.k8s.io/etcd:3.6.4-0
docker tag registry.aliyuncs.com/google_containers/pause:3.10.1 registry.k8s.io/pause:3.10.1
docker tag registry.aliyuncs.com/google_containers/storage-provisioner:v5 gcr.io/k8s-minikube/storage-provisioner:v5

# 4. 启动 minikube（如果内存不足，使用 --force 参数）
minikube delete  # 先删除现有实例
minikube start --memory=1500mb --cpus=2 --force

# 5. 登录 GitHub Container Registry
docker login ghcr.io

# 6. 修改镜像地址（替换 YOUR_GITHUB_USERNAME）
sed -i 's/ghcr.io\/OWNER\/hello-go:latest/ghcr.io\/YOUR_GITHUB_USERNAME\/hello-go:latest/g' k8s/deployment.yaml

# 7. 构建镜像
docker build -t ghcr.io/YOUR_GITHUB_USERNAME/hello-go:latest .

# 8. 推送镜像
docker push ghcr.io/YOUR_GITHUB_USERNAME/hello-go:latest

# 9. 部署应用
kubectl apply -f k8s/deployment.yaml

# 10. 等待就绪
kubectl wait --for=condition=ready pod -l app=hello --timeout=60s

# 11. 访问应用
minikube ip
curl http://$(minikube ip):30080
```

