package main

import (
	"github.com/spf13/cobra"
)

// kubeconfigFlag, namespaceFlag, and releaseNameFlag are persistent flags
// shared by every subcommand that talks to a cluster.
var (
	kubeconfigFlag  string
	namespaceFlag   string
	releaseNameFlag string
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "miramar",
		Short: "Deploy and validate an OpenAI-compatible model endpoint on an existing Kubernetes cluster",
		Long: `miramar deploys and validates one model-serving endpoint on a GPU-enabled
Kubernetes cluster you already have access to. It does not provision clusters,
GPU drivers, or node pools.

See ROADMAP.md for what's implemented today vs. planned.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&kubeconfigFlag, "kubeconfig", "", "Path to kubeconfig (default: $KUBECONFIG or ~/.kube/config)")
	cmd.PersistentFlags().StringVarP(&namespaceFlag, "namespace", "n", "miramar-ai-platform", "Kubernetes namespace for the deployment")
	cmd.PersistentFlags().StringVar(&releaseNameFlag, "release-name", "miramar", "Helm release name")

	cmd.AddCommand(
		newDoctorCmd(),
		newNewCmd(),
		newDeployCmd(),
		newValidateCmd(),
		newUndeployCmd(),
	)

	return cmd
}
