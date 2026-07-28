// Package validator confirms a deployed endpoint is actually healthy and
// serving requests. It does its own independent readiness wait rather than
// trusting the deployer's exit code, and reaches the ClusterIP endpoint via a
// client-go SPDY port-forward rather than shelling out to kubectl.
package validator

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/miramar-labs-org/miramar-ai-platform-common/k8sclient"
)

// Options configures one validation run.
type Options struct {
	KubeConfig      string
	KubeContext     string
	Namespace       string
	ReleaseName     string
	ServedModelName string
	MaxTokens       int
	Timeout         time.Duration
}

// Result reports what validation observed.
type Result struct {
	Models      []string
	PromptsOK   int
	PromptsFail int
	Failures    []string
}

// smokePrompts are short, well-formedness-only checks — they assert a
// non-empty response, not any particular content.
var smokePrompts = []string{
	"Say hello in one short sentence.",
	"What is 2 + 2?",
	"Name one primary color.",
}

// Run waits for a ready pod, port-forwards to it, and exercises
// /v1/models and /v1/chat/completions. It returns a non-nil error if
// anything failed, alongside a Result describing what did and didn't work.
func Run(ctx context.Context, opts Options) (*Result, error) {
	flags := k8sclient.ConfigFlags(opts.KubeConfig, opts.KubeContext, opts.Namespace)
	restConfig, err := flags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("resolving kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("building kubernetes client: %w", err)
	}

	pod, err := waitForReadyPod(ctx, clientset, opts.Namespace, opts.ReleaseName, opts.Timeout)
	if err != nil {
		return nil, err
	}

	localPort, stop, err := portForward(restConfig, clientset, opts.Namespace, pod.Name, 8000)
	if err != nil {
		return nil, fmt.Errorf("port-forwarding to pod %q: %w", pod.Name, err)
	}
	defer stop()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", localPort)

	models, err := fetchModels(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GET /v1/models: %w", err)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("no models reported by /v1/models")
	}

	result := &Result{Models: models}
	servedName := opts.ServedModelName
	if servedName == "" {
		servedName = models[0]
	}

	for _, prompt := range smokePrompts {
		content, err := chatCompletion(ctx, baseURL, servedName, prompt, opts.MaxTokens)
		if err != nil || content == "" {
			result.PromptsFail++
			result.Failures = append(result.Failures, fmt.Sprintf("%q: %v", prompt, err))
			continue
		}
		result.PromptsOK++
	}

	if result.PromptsFail > 0 {
		return result, fmt.Errorf("%d/%d smoke-test prompts failed", result.PromptsFail, len(smokePrompts))
	}
	return result, nil
}

// waitForReadyPod polls for a Running+Ready pod matching the chart's static
// selector labels — never trusts the deployer's own wait/exit code.
func waitForReadyPod(ctx context.Context, clientset kubernetes.Interface, namespace, releaseName string, timeout time.Duration) (*corev1.Pod, error) {
	selector := fmt.Sprintf("app.kubernetes.io/name=serving-vllm,app.kubernetes.io/instance=%s", releaseName)

	var found *corev1.Pod
	err := wait.PollUntilContextTimeout(ctx, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return false, err
		}
		for i := range pods.Items {
			if pods.Items[i].Status.Phase == corev1.PodRunning && isPodReady(&pods.Items[i]) {
				found = &pods.Items[i]
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return nil, fmt.Errorf("waiting for a ready pod (selector %q): %w", selector, err)
	}
	return found, nil
}

func isPodReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}
