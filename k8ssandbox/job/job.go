package job

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func RunPythonInK8sJob(pythonCode string) (string, error) {
	// 创建 K8s 客户端
	// config, err := rest.InClusterConfig() // 如果在集群内运行
	// if err != nil {
	// 	// 如果在集群外运行，使用 kubeconfig
	// 	// config, err = clientcmd.BuildConfigFromFlags("", "/path/to/kubeconfig")
	// 	return "", fmt.Errorf("failed to get in-cluster config: %w", err)
	// }

	// 或者指定特定的 kubeconfig 文件路径
	config, err := clientcmd.BuildConfigFromFlags("https://49.232.226.50:6443", "/home/skf/data/workcode/godemo/k8ssandbox/conf/k8s.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to build config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create k8s client: %w", err)
	}

	ctx := context.Background()

	jobName := fmt.Sprintf("python-sandbox-%d", time.Now().UnixNano())

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "default", // 确保命名空间存在
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "python-runner",
							Image:   "registry.yygu.cn/base/python:3.10",
							Command: []string{"python", "-c", pythonCode},
							SecurityContext: &v1.SecurityContext{
								RunAsNonRoot:             &[]bool{true}[0],
								RunAsUser:                &[]int64{1000}[0],
								RunAsGroup:               &[]int64{1000}[0],
								ReadOnlyRootFilesystem:   &[]bool{true}[0],
								AllowPrivilegeEscalation: &[]bool{false}[0],
								Capabilities: &v1.Capabilities{
									Drop: []v1.Capability{"ALL"},
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &[]int32{0}[0],
			TTLSecondsAfterFinished: &[]int32{30}[0],
		},
	}

	createdJob, err := clientset.BatchV1().Jobs("default").Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	fmt.Printf("Created job: %s\n", createdJob.Name)

	// 等待 Job 完成
	for {
		job, err := clientset.BatchV1().Jobs("default").Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to get job: %w", err)
		}

		if job.Status.Succeeded > 0 {
			break
		}
		if job.Status.Failed > 0 {
			return "", fmt.Errorf("job failed")
		}
		time.Sleep(1 * time.Second)
	}

	// 获取日志
	pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for job")
	}

	podName := pods.Items[0].Name
	req := clientset.CoreV1().Pods("default").GetLogs(podName, &v1.PodLogOptions{})
	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	return string(logs), nil
}
