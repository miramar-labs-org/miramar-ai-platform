// Package k8sclient provides shared kubeconfig resolution for the deployer
// and validator packages, so both Helm's REST client and a plain client-go
// clientset agree on which cluster/namespace they're talking to.
package k8sclient

import "k8s.io/cli-runtime/pkg/genericclioptions"

// ConfigFlags builds kubeconfig resolution shared by Helm's
// action.RESTClientGetter and a plain client-go rest.Config/clientset.
// kubeconfigPath and kubeContext may be empty, in which case the standard
// --kubeconfig > $KUBECONFIG > ~/.kube/config precedence applies.
func ConfigFlags(kubeconfigPath, kubeContext, namespace string) *genericclioptions.ConfigFlags {
	flags := genericclioptions.NewConfigFlags(false)
	if kubeconfigPath != "" {
		flags.KubeConfig = &kubeconfigPath
	}
	if kubeContext != "" {
		flags.Context = &kubeContext
	}
	flags.Namespace = &namespace
	return flags
}
