package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/miramar-labs-org/miramar-ai-platform/internal/validator"
)

func newValidateCmd() *cobra.Command {
	var (
		servedModelName string
		maxTokens       int
		timeout         time.Duration
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Confirm the deployed endpoint is healthy and serving requests",
		Long: `validate waits for a ready pod, opens its own port-forward to it (it never
trusts "deploy"'s exit code), then checks /v1/models and sends a small set of
smoke-test prompts to /v1/chat/completions.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := validator.Run(cmd.Context(), validator.Options{
				KubeConfig:      kubeconfigFlag,
				Namespace:       namespaceFlag,
				ReleaseName:     releaseNameFlag,
				ServedModelName: servedModelName,
				MaxTokens:       maxTokens,
				Timeout:         timeout,
			})

			out := cmd.OutOrStdout()
			if result != nil {
				fmt.Fprintf(out, "Models: %v\n", result.Models)
				fmt.Fprintf(out, "Smoke-test prompts: %d ok, %d failed\n", result.PromptsOK, result.PromptsFail)
				for _, f := range result.Failures {
					fmt.Fprintln(out, styleFail(fmt.Sprintf("  FAILED: %s", f)))
				}
			}
			if err != nil {
				return fmt.Errorf("miramar validate: %w", err)
			}
			fmt.Fprintln(out, styleOK("Endpoint is healthy and serving requests."))
			return nil
		},
	}

	cmd.Flags().StringVar(&servedModelName, "served-model-name", "", "Model name to request (default: first model reported by /v1/models)")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 32, "max_tokens for each smoke-test request")
	cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Minute, "How long to wait for a ready pod before giving up")

	return cmd
}
