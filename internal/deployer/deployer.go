// Package deployer installs, upgrades, and uninstalls a release via the
// native Helm SDK. It knows nothing about where a chart came from — the
// template package handles resolving/copying charts; this package only
// consumes an already-loaded *chart.Chart.
package deployer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/miramar-labs-org/miramar-ai-platform/internal/k8sclient"
)

// ClientOptions configures kubeconfig/cluster resolution, shared by every
// operation a Client performs.
type ClientOptions struct {
	KubeConfig  string
	KubeContext string
	Namespace   string
}

// Client wraps a Helm action.Configuration bound to one cluster/namespace.
type Client struct {
	cfg        *action.Configuration
	namespace  string
	restConfig *rest.Config
}

// NewClient resolves the kubeconfig and initializes the Helm SDK against it.
func NewClient(opts ClientOptions) (*Client, error) {
	flags := k8sclient.ConfigFlags(opts.KubeConfig, opts.KubeContext, opts.Namespace)

	cfg := new(action.Configuration)
	debugLog := func(format string, v ...interface{}) {}
	if err := cfg.Init(flags, opts.Namespace, os.Getenv("HELM_DRIVER"), debugLog); err != nil {
		return nil, fmt.Errorf("initializing helm: %w", err)
	}

	restConfig, err := flags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("resolving kubeconfig: %w", err)
	}

	return &Client{cfg: cfg, namespace: opts.Namespace, restConfig: restConfig}, nil
}

// DeployOptions configures one Deploy call.
type DeployOptions struct {
	ReleaseName string
	Wait        bool
	Timeout     time.Duration
}

// Deploy installs chrt as a new release, or upgrades it in place if a release
// by that name already exists — Helm's own "upgrade --install" idiom. The
// namespace is created automatically on first install.
func (c *Client) Deploy(ctx context.Context, chrt *chart.Chart, vals map[string]interface{}, opts DeployOptions) (*release.Release, error) {
	hist := action.NewHistory(c.cfg)
	hist.Max = 1
	_, err := hist.Run(opts.ReleaseName)

	switch {
	case errors.Is(err, driver.ErrReleaseNotFound):
		install := action.NewInstall(c.cfg)
		install.ReleaseName = opts.ReleaseName
		install.Namespace = c.namespace
		install.CreateNamespace = true
		install.Wait = opts.Wait
		install.Timeout = opts.Timeout
		return install.RunWithContext(ctx, chrt, vals)
	case err != nil:
		return nil, fmt.Errorf("checking release history for %q: %w", opts.ReleaseName, err)
	default:
		upgrade := action.NewUpgrade(c.cfg)
		upgrade.Namespace = c.namespace
		upgrade.Wait = opts.Wait
		upgrade.Timeout = opts.Timeout
		return upgrade.RunWithContext(ctx, opts.ReleaseName, chrt, vals)
	}
}

// Uninstall removes the Helm release. A missing release is not an error. If
// purgeNamespace is set, the namespace itself is deleted afterward.
func (c *Client) Uninstall(ctx context.Context, releaseName string, purgeNamespace bool) error {
	uninstall := action.NewUninstall(c.cfg)
	if _, err := uninstall.Run(releaseName); err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		return fmt.Errorf("uninstalling release %q: %w", releaseName, err)
	}

	if !purgeNamespace {
		return nil
	}

	clientset, err := kubernetes.NewForConfig(c.restConfig)
	if err != nil {
		return fmt.Errorf("building kubernetes client: %w", err)
	}
	if err := clientset.CoreV1().Namespaces().Delete(ctx, c.namespace, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("deleting namespace %q: %w", c.namespace, err)
	}
	return nil
}
