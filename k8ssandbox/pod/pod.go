package pod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type SandboxPod struct {
	Name string
	IP   string
	Busy bool
}

type SandboxManager struct {
	clientset *kubernetes.Clientset
	namespace string
	pods      []*SandboxPod
	mu        sync.Mutex
}

// 初始化 client-go
func NewSandboxManager(namespace string) (*SandboxManager, error) {
	// config, _ := rest.InClusterConfig()
	// 或者指定特定的 kubeconfig 文件路径
	config, err := clientcmd.BuildConfigFromFlags("https://49.232.226.50:6443", "/home/skf/data/workcode/godemo/k8ssandbox/conf/k8s.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}
	clientset, _ := kubernetes.NewForConfig(config)
	return &SandboxManager{clientset: clientset, namespace: namespace}, err
}

// 启动多个 Python 沙箱 Pod
func (m *SandboxManager) InitPods(count int) error {
	for i := 0; i < count; i++ {
		// name := fmt.Sprintf("py-sandbox-%d", i)
		// pod := &corev1.Pod{
		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Name:      name,
		// 		Namespace: m.namespace,
		// 		Labels:    map[string]string{"app": "py-sandbox"},
		// 	},
		// 	Spec: corev1.PodSpec{
		// 		RestartPolicy: corev1.RestartPolicyAlways,
		// 		Containers: []corev1.Container{
		// 			{
		// 				Name:      "python",
		// 				Image:     "registry.yygu.cn/skftest/python-sandbox:v0.0.1",
		// 				Ports:     []corev1.ContainerPort{{ContainerPort: 8080}},
		// 				Resources: corev1.ResourceRequirements{},
		// 			},
		// 		},
		// 	},
		// }
		// a, err := m.clientset.CoreV1().Pods(m.namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
		// if err != nil {
		// 	return err
		// }
		m.pods = append(m.pods, &SandboxPod{Name: "py-sandbox-0", IP: "49.232.226.50:30088", Busy: false})
	}
	return nil
}

// 获取空闲 Pod
func (m *SandboxManager) GetIdlePod() *SandboxPod {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.pods {
		if !p.Busy {
			p.Busy = true
			return p
		}
	}
	return nil
}

// 执行 Python 代码
func (m *SandboxManager) RunCode(code string) (string, error) {
	pod := m.GetIdlePod()
	if pod == nil {
		return "", fmt.Errorf("no idle pod available")
	}
	defer func() {
		m.mu.Lock()
		pod.Busy = false
		m.mu.Unlock()
	}()

	payload := map[string]string{"code": code}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("http://%s/run", pod.IP)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	resBody, _ := io.ReadAll(resp.Body)
	return string(resBody), nil
}
