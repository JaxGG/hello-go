# Kubernetes 配置文件说明

## 文件列表

### 必需文件

- **deployment.yaml** - 包含 Deployment 和 Service
  - Deployment: 应用部署配置
  - Service: 服务暴露配置（NodePort）

### 可选文件

- **ingress.yaml** - Ingress 配置（需要 Ingress Controller）
  - 用于从集群外部通过域名访问
  - 需要修改 `host` 字段为你的域名
  - 需要安装 Ingress Controller（如 nginx-ingress）

- **hpa.yaml** - 水平自动扩缩容配置
  - 根据 CPU/内存使用率自动调整 Pod 数量
  - 需要安装 metrics-server

## 部署方式

### 基础部署（仅 Deployment + Service）

```bash
kubectl apply -f k8s/deployment.yaml
```

### 完整部署（包含 Ingress 和 HPA）

```bash
# 确保已安装 metrics-server（HPA 需要）
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# 部署所有资源
kubectl apply -f k8s/
```

## 访问方式

### 方式 1: NodePort（通过 IP 访问，推荐）

```bash
# 获取节点 IP
kubectl get nodes -o wide

# 通过节点 IP + 30080 端口访问
curl http://<NODE_IP>:30080

# 或在浏览器访问
# http://<NODE_IP>:30080
```

**示例**：
- 如果节点 IP 是 `192.168.1.100`，访问地址为：`http://192.168.1.100:30080`

### 方式 2: LoadBalancer（云平台，如 AWS/GCP/Azure）

如果使用云平台，可以使用 `service-loadbalancer.yaml`：

```bash
# 部署 LoadBalancer Service
kubectl apply -f k8s/service-loadbalancer.yaml

# 查看外部 IP
kubectl get svc hello-lb

# 等待 EXTERNAL-IP 分配后，直接访问该 IP
curl http://<EXTERNAL-IP>
```

### 方式 3: Port Forward（开发测试）

```bash
kubectl port-forward svc/hello 8080:8080
# 访问 http://localhost:8080
```

### 方式 4: Ingress（需要域名）

1. 安装 Ingress Controller（如 nginx-ingress）
2. 修改 `ingress.yaml` 中的域名
3. 配置 DNS 解析
4. 访问 `http://hello.example.com`

