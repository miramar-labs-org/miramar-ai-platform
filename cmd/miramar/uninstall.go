package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/miramar-labs-org/miramar-ai-platform/internal/deployer"
)

func newUninstallCmd() *cobra.Command {
	var (
		purgeNamespace bool
		timeout        time.Duration
	)

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Tear the deployment back down",
		Long: `uninstall removes the Helm release from the target namespace. By default the
namespace itself is left in place; pass --purge-namespace to delete it too.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deployer.NewClient(deployer.ClientOptions{
				KubeConfig: kubeconfigFlag,
				Namespace:  namespaceFlag,
			})
			if err != nil {
				return fmt.Errorf("miramar uninstall: %w", err)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			if err := client.Uninstall(ctx, releaseNameFlag, purgeNamespace); err != nil {
				return fmt.Errorf("miramar uninstall: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, styleOK(fmt.Sprintf("Uninstalled release %q from namespace %q", releaseNameFlag, namespaceFlag)))
			if purgeNamespace {
				fmt.Fprintln(out, styleOK(fmt.Sprintf("Namespace %q deleted", namespaceFlag)))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&purgeNamespace, "purge-namespace", false, "Also delete the namespace after uninstalling the release")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "How long to wait for uninstall/namespace deletion to finish")

	return cmd
}
