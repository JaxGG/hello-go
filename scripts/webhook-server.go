package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type WebhookPayload struct {
	Image  string `json:"image"`
	Tag    string `json:"tag"`
	Ref    string `json:"ref"`
	Commit string `json:"commit"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var payload WebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Received deployment request: Image=%s, Tag=%s, Commit=%s", payload.Image, payload.Tag, payload.Commit)

		// 更新部署
		if err := updateDeployment(payload.Image); err != nil {
			log.Printf("Error updating deployment: %v", err)
			http.Error(w, fmt.Sprintf("Deployment failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Deployment updated",
			"image":   payload.Image,
		})

		log.Printf("Deployment updated successfully: %s", payload.Image)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Webhook server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func updateDeployment(image string) error {
	// 查找项目根目录（包含 k8s/deployment.yaml 的目录）
	workDir, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	log.Printf("Using project root: %s", workDir)

	// 从远程镜像名提取本地镜像名（用于 kind）
	// 例如: ghcr.io/username/hello-go:latest -> hello-go:latest
	localImageName := extractLocalImageName(image)

	log.Printf("Pulling image: %s", image)
	// 拉取远程镜像到本地
	cmd := exec.Command("docker", "pull", image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker pull failed: %v, output: %s", err, string(output))
	}
	log.Printf("Image pulled successfully: %s", string(output))

	// 给镜像打本地标签
	log.Printf("Tagging image as: %s", localImageName)
	cmd = exec.Command("docker", "tag", image, localImageName)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker tag failed: %v, output: %s", err, string(output))
	}
	log.Printf("Image tagged successfully")

	// 获取 kind 集群名称（默认 hello-go-cluster）
	kindClusterName := os.Getenv("KIND_CLUSTER_NAME")
	if kindClusterName == "" {
		kindClusterName = "hello-go-cluster"
	}

	// 将镜像加载到 kind 集群
	log.Printf("Loading image into kind cluster: %s", kindClusterName)
	cmd = exec.Command("kind", "load", "docker-image", localImageName, "--name", kindClusterName)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kind load docker-image failed: %v, output: %s", err, string(output))
	}
	log.Printf("Image loaded into kind cluster: %s", string(output))

	// 使用 kubectl set image 更新部署（使用本地镜像名）
	cmd = exec.Command("kubectl", "set", "image", "deployment/hello", fmt.Sprintf("hello=%s", localImageName))
	cmd.Dir = workDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl set image failed: %v, output: %s", err, string(output))
	}

	log.Printf("Deployment image updated: %s", string(output))

	// 等待部署完成
	cmd = exec.Command("kubectl", "rollout", "status", "deployment/hello", "--timeout=60s")
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: rollout status check failed: %v", err)
		// 不返回错误，因为镜像已经更新
	}

	return nil
}

// findProjectRoot 查找项目根目录（包含 k8s/deployment.yaml 的目录）
func findProjectRoot() (string, error) {
	// 从当前工作目录开始向上查找
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		deploymentPath := filepath.Join(dir, "k8s", "deployment.yaml")
		if _, err := os.Stat(deploymentPath); err == nil {
			return dir, nil
		}

		// 向上查找
		parent := filepath.Dir(dir)
		if parent == dir {
			// 已经到达根目录
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found (k8s/deployment.yaml not found)")
}

// extractLocalImageName 从完整镜像名提取本地镜像名
// 例如: ghcr.io/username/hello-go:latest -> hello-go:latest
func extractLocalImageName(fullImageName string) string {
	// 移除 registry 前缀
	parts := strings.Split(fullImageName, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// 如果包含冒号，保留标签；否则添加 :latest
		if strings.Contains(lastPart, ":") {
			return lastPart
		}
		return lastPart + ":latest"
	}
	return fullImageName
}
