// Package doctor inspects an existing Kubernetes cluster's readiness for
// "miramar deploy" without modifying it. Every check asks "can I talk to
// Kubernetes, can I schedule a GPU pod, do the prerequisites exist" — never
// "is this cluster distribution supported" (see ADR-007). Checks report
// PASS, WARN, or FAIL; only a FAIL should block a user from proceeding. Only
// API connectivity, schedulable GPU, and the two deploy-path authorization
// checks (target namespace access, Helm release-storage access) can FAIL —
// storage-class, ingress-controller, and observability-integration discovery
// are informational only and report Warn at worst, even on a list error.
package doctor

import (
	"context"
	"fmt"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	"github.com/miramar-labs-org/miramar-ai-platform/internal/k8sclient"
)

// Status is the outcome of one check.
type Status int

const (
	Pass Status = iota
	Warn
	Fail
)

func (s Status) String() string {
	switch s {
	case Pass:
		return "PASS"
	case Warn:
		return "WARN"
	case Fail:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// CheckResult reports the outcome of one prerequisite check. Remediation is
// set only when Status is Warn or Fail.
type CheckResult struct {
	Name        string
	Status      Status
	Detail      string
	Remediation string
}

// Report is the full set of check results from one doctor run.
type Report struct {
	Results []CheckResult
}

// HasFailures reports whether any check in the report failed.
func (r *Report) HasFailures() bool {
	for _, res := range r.Results {
		if res.Status == Fail {
			return true
		}
	}
	return false
}

// Options configures one doctor run.
type Options struct {
	KubeConfig  string
	KubeContext string
	Namespace   string
}

// Run resolves the active kubeconfig and runs every check against it. It
// never modifies the cluster. A non-nil error means the cluster connection
// itself could not be established, so no checks could run at all —
// individual check failures are reported in the returned Report, not as an
// error.
func Run(ctx context.Context, opts Options) (*Report, error) {
	flags := k8sclient.ConfigFlags(opts.KubeConfig, opts.KubeContext, opts.Namespace)
	restConfig, err := flags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("resolving kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("building kubernetes client: %w", err)
	}

	report := &Report{}
	report.Results = append(report.Results, checkAPIConnectivity(clientset))
	report.Results = append(report.Results, checkDistribution(clientset))
	report.Results = append(report.Results, checkSchedulableGPU(ctx, clientset))
	report.Results = append(report.Results, checkStorageClasses(ctx, clientset))
	report.Results = append(report.Results, checkIngressController(ctx, clientset))
	report.Results = append(report.Results, checkObservabilityStack(clientset))

	nsResult, namespaceMissing := checkNamespace(ctx, clientset, opts.Namespace)
	report.Results = append(report.Results, nsResult)
	report.Results = append(report.Results, checkHelmReleaseStorage(flags, opts.Namespace, namespaceMissing))

	return report, nil
}

// checkAPIConnectivity confirms the current kubeconfig context actually
// reaches a Kubernetes API server, and reports which one.
func checkAPIConnectivity(clientset kubernetes.Interface) CheckResult {
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return CheckResult{
			Name:        "Kubernetes API connectivity",
			Status:      Fail,
			Detail:      fmt.Sprintf("could not reach the Kubernetes API: %v", err),
			Remediation: "confirm the current kubectl context points at a reachable cluster (kubectl cluster-info), or pass --kubeconfig",
		}
	}
	return CheckResult{
		Name:   "Kubernetes API connectivity",
		Status: Pass,
		Detail: fmt.Sprintf("reachable, server version %s", version.GitVersion),
	}
}

// checkDistribution makes a best-effort, informational-only guess at the
// cluster distribution from its reported version string. It never fails —
// deploy only needs a reachable API and a schedulable GPU (ADR-007), so an
// unrecognized distribution is not a problem to report as one.
func checkDistribution(clientset kubernetes.Interface) CheckResult {
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return CheckResult{
			Name:   "Cluster distribution (informational)",
			Status: Pass,
			Detail: "could not be determined (API unreachable — see Kubernetes API connectivity check)",
		}
	}

	gitVersion := version.GitVersion
	switch {
	case strings.Contains(gitVersion, "+k3s"):
		return CheckResult{Name: "Cluster distribution (informational)", Status: Pass, Detail: fmt.Sprintf("k3s (%s)", gitVersion)}
	case strings.Contains(gitVersion, "-gke."):
		return CheckResult{Name: "Cluster distribution (informational)", Status: Pass, Detail: fmt.Sprintf("GKE (%s)", gitVersion)}
	case strings.Contains(gitVersion, "-eks-"):
		return CheckResult{Name: "Cluster distribution (informational)", Status: Pass, Detail: fmt.Sprintf("EKS (%s)", gitVersion)}
	default:
		return CheckResult{
			Name:   "Cluster distribution (informational)",
			Status: Pass,
			Detail: fmt.Sprintf("not reliably identified from version string %q — compatible by design if other checks pass (ADR-007)", gitVersion),
		}
	}
}

// checkSchedulableGPU confirms at least one node advertises an allocatable
// nvidia.com/gpu resource, which is what every serving template requests.
func checkSchedulableGPU(ctx context.Context, clientset kubernetes.Interface) CheckResult {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return CheckResult{
			Name:        "Schedulable GPU",
			Status:      Fail,
			Detail:      fmt.Sprintf("could not list nodes: %v", err),
			Remediation: "confirm the current user/context has permission to list nodes",
		}
	}

	var gpuNodes int
	var totalGPUs int64
	for _, node := range nodes.Items {
		qty, ok := node.Status.Allocatable[corev1.ResourceName("nvidia.com/gpu")]
		if !ok {
			continue
		}
		if n, ok := qty.AsInt64(); ok && n > 0 {
			gpuNodes++
			totalGPUs += n
		}
	}

	if totalGPUs == 0 {
		return CheckResult{
			Name:        "Schedulable GPU",
			Status:      Fail,
			Detail:      fmt.Sprintf("no node advertises an allocatable nvidia.com/gpu resource (%d nodes checked)", len(nodes.Items)),
			Remediation: "install/verify the NVIDIA GPU Operator or device plugin on at least one node",
		}
	}
	return CheckResult{
		Name:   "Schedulable GPU",
		Status: Pass,
		Detail: fmt.Sprintf("%d allocatable GPU(s) across %d node(s)", totalGPUs, gpuNodes),
	}
}

// checkStorageClasses discovers what StorageClasses exist on the cluster.
// This is informational only, never Fail — even a list error is reported as
// Warn, since no template in this repo declares a storageClassName yet (see
// docs/CURRENT_STATE.md), so nothing actually requires one. It exists so
// `doctor` demonstrates platform awareness ahead of a template that does
// need one.
func checkStorageClasses(ctx context.Context, clientset kubernetes.Interface) CheckResult {
	classes, err := clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return CheckResult{
			Name:        "Storage classes",
			Status:      Warn,
			Detail:      fmt.Sprintf("could not list storage classes: %v", err),
			Remediation: "confirm the current user/context has permission to list storageclasses.storage.k8s.io; not required for the default vLLM template today",
		}
	}
	if len(classes.Items) == 0 {
		return CheckResult{
			Name:        "Storage classes",
			Status:      Warn,
			Detail:      "no StorageClass found",
			Remediation: "not required for the default vLLM template today; install a StorageClass via your CSI driver if a future template requests persistent storage",
		}
	}

	defaultClass := ""
	for _, sc := range classes.Items {
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			defaultClass = sc.Name
			break
		}
	}
	detail := fmt.Sprintf("%d storage class(es) found", len(classes.Items))
	if defaultClass != "" {
		detail += fmt.Sprintf(", default: %q", defaultClass)
	} else {
		detail += ", no default class set"
	}
	return CheckResult{Name: "Storage classes", Status: Pass, Detail: detail}
}

// checkIngressController discovers whether any IngressClass is registered.
// Informational only — the golden path doesn't expose an Ingress today, so
// even a list error is reported as Warn, not Fail.
func checkIngressController(ctx context.Context, clientset kubernetes.Interface) CheckResult {
	classes, err := clientset.NetworkingV1().IngressClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return CheckResult{
			Name:        "Ingress controller",
			Status:      Warn,
			Detail:      fmt.Sprintf("could not list ingress classes: %v", err),
			Remediation: "confirm the current user/context has permission to list ingressclasses.networking.k8s.io; not required for the in-cluster golden path",
		}
	}
	if len(classes.Items) == 0 {
		return CheckResult{
			Name:        "Ingress controller",
			Status:      Warn,
			Detail:      "no IngressClass found",
			Remediation: "not required for the in-cluster golden path; install an ingress controller if you plan to expose an endpoint outside the cluster",
		}
	}

	names := make([]string, 0, len(classes.Items))
	for _, ic := range classes.Items {
		names = append(names, ic.Name)
	}
	return CheckResult{Name: "Ingress controller", Status: Pass, Detail: fmt.Sprintf("IngressClass(es) found: %s", strings.Join(names, ", "))}
}

// observabilityGroupVersions maps a discoverable API group/version to the
// human-readable integration it indicates is installed. Order is fixed so
// Detail text is stable across runs.
var observabilityGroupVersions = []struct {
	groupVersion string
	label        string
}{
	{"metrics.k8s.io/v1beta1", "metrics-server"},
	{"monitoring.coreos.com/v1", "Prometheus Operator"},
	{"opentelemetry.io/v1alpha1", "OpenTelemetry Operator"},
}

// checkObservabilityStack discovers standard observability integrations the
// user may already run, per ADR-016: CE detects and reports these as
// informational/warning-level context, but never fails on their absence —
// Miramar does not require, install, or manage an observability stack.
func checkObservabilityStack(clientset kubernetes.Interface) CheckResult {
	var found []string
	for _, gv := range observabilityGroupVersions {
		if _, err := clientset.Discovery().ServerResourcesForGroupVersion(gv.groupVersion); err == nil {
			found = append(found, gv.label)
		}
	}

	if len(found) == 0 {
		return CheckResult{
			Name:        "Observability integrations",
			Status:      Warn,
			Detail:      "no standard observability integrations detected (metrics-server, Prometheus Operator, OpenTelemetry Operator)",
			Remediation: "optional — see ADR-016; Miramar does not require an observability stack",
		}
	}
	return CheckResult{Name: "Observability integrations", Status: Pass, Detail: fmt.Sprintf("detected: %s", strings.Join(found, ", "))}
}

// checkNamespace reports whether the target namespace already exists. A
// missing namespace is a WARN, not a FAIL — deploy creates it automatically.
// The second return value is true only when the namespace is confirmed not
// to exist yet, so callers (checkHelmReleaseStorage) can skip checks that
// are meaningless against a namespace that doesn't exist.
func checkNamespace(ctx context.Context, clientset kubernetes.Interface, namespace string) (CheckResult, bool) {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	switch {
	case err == nil:
		return CheckResult{
			Name:   "Target namespace",
			Status: Pass,
			Detail: fmt.Sprintf("namespace %q already exists", namespace),
		}, false
	case apierrors.IsNotFound(err):
		return CheckResult{
			Name:        "Target namespace",
			Status:      Warn,
			Detail:      fmt.Sprintf("namespace %q does not exist yet", namespace),
			Remediation: "no action needed — `miramar deploy` creates it automatically",
		}, true
	default:
		return CheckResult{
			Name:        "Target namespace",
			Status:      Fail,
			Detail:      fmt.Sprintf("could not check namespace %q: %v", namespace, err),
			Remediation: "confirm the current user/context has permission to get namespaces",
		}, false
	}
}

// checkHelmReleaseStorage exercises the same Helm SDK path deployer.Client
// uses (an action.Configuration against the target namespace) with a
// read-only release list, catching RBAC gaps on Helm's release-storage
// Secrets that a plain Kubernetes API connectivity check would not.
//
// If the target namespace is confirmed not to exist yet, this check is
// skipped rather than run for real: many clusters grant namespace-scoped
// RBAC only once the namespace exists, so listing Secrets in a namespace
// that isn't there yet can spuriously fail even though `deploy` (which
// creates the namespace itself) would work fine. Skipping keeps a
// fresh-cluster `doctor` run from hard-failing before `deploy` ever gets a
// chance to create the namespace.
func checkHelmReleaseStorage(flags *genericclioptions.ConfigFlags, namespace string, namespaceMissing bool) CheckResult {
	if namespaceMissing {
		return CheckResult{
			Name:        "Helm release storage access",
			Status:      Warn,
			Detail:      fmt.Sprintf("not checked — namespace %q does not exist yet", namespace),
			Remediation: "no action needed — `miramar deploy` creates the namespace and its own Helm release storage automatically",
		}
	}

	cfg := new(action.Configuration)
	debugLog := func(format string, v ...interface{}) {}
	if err := cfg.Init(flags, namespace, os.Getenv("HELM_DRIVER"), debugLog); err != nil {
		return CheckResult{
			Name:        "Helm release storage access",
			Status:      Fail,
			Detail:      fmt.Sprintf("could not initialize Helm: %v", err),
			Remediation: "confirm the current user/context has permission to access Secrets in the target namespace",
		}
	}

	list := action.NewList(cfg)
	list.All = true
	releases, err := list.Run()
	if err != nil {
		return CheckResult{
			Name:        "Helm release storage access",
			Status:      Fail,
			Detail:      fmt.Sprintf("could not list Helm releases in namespace %q: %v", namespace, err),
			Remediation: "confirm the current user/context has get/list/watch permission on Secrets in the target namespace",
		}
	}
	return CheckResult{
		Name:   "Helm release storage access",
		Status: Pass,
		Detail: fmt.Sprintf("%d existing release(s) visible in namespace %q", len(releases), namespace),
	}
}
