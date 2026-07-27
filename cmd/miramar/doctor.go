package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/miramar-labs-org/miramar-ai-platform/internal/doctor"
)

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check cluster prerequisites without modifying anything",
		Long: `doctor inspects the active kubeconfig context and reports whether the
cluster is usable for "miramar deploy" — API connectivity, a best-effort
distribution guess (informational only), a schedulable GPU, storage classes,
an ingress controller, standard observability integrations (metrics-server,
Prometheus Operator, OpenTelemetry Operator), the target namespace, and
Helm's access to its own release storage.

It is read-only: it never provisions, modifies, or fixes anything. Only a
FAIL causes a non-zero exit; WARN is informational. Only API connectivity,
schedulable GPU, target namespace access, and Helm release-storage access
can FAIL — storage-class, ingress-controller, and observability discovery
are informational only. If the target namespace doesn't exist yet, Helm
release-storage access is skipped (WARN, not FAIL) rather than checked
against a namespace "deploy" hasn't created yet.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := doctor.Run(cmd.Context(), doctor.Options{
				KubeConfig: kubeconfigFlag,
				Namespace:  namespaceFlag,
			})
			if err != nil {
				return fmt.Errorf("miramar doctor: %w", err)
			}

			out := cmd.OutOrStdout()
			for _, result := range report.Results {
				fmt.Fprintf(out, "%s %s: %s\n", styleDoctorStatus(result.Status), result.Name, result.Detail)
				if result.Remediation != "" {
					fmt.Fprintf(out, "       -> %s\n", result.Remediation)
				}
			}

			if report.HasFailures() {
				return fmt.Errorf("miramar doctor: one or more checks failed")
			}
			return nil
		},
	}

	return cmd
}
