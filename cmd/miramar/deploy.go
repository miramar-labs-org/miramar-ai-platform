package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/miramar-labs-org/miramar-ai-platform-common/deployer"
	"github.com/miramar-labs-org/miramar-ai-platform/internal/template"
)

func newDeployCmd() *cobra.Command {
	var (
		templateType    string
		chartDir        string
		valuesFile      string
		model           string
		servedModelName string
		waitTimeout     time.Duration
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy one model endpoint via a Helm chart",
		Long: `deploy installs (or upgrades, if a release by this name already exists) one
model-serving endpoint into the target namespace via the native Helm SDK.

By default it deploys the built-in "serving-vllm" template
(Qwen/Qwen2.5-1.5B-Instruct). Use --chart-dir to deploy from a directory
produced by "miramar new" instead.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if chartDir != "" && cmd.Flags().Changed("type") {
				return fmt.Errorf("miramar deploy: --type and --chart-dir are mutually exclusive")
			}

			var chrt *chart.Chart
			var err error
			if chartDir != "" {
				chrt, err = template.LoadDir(chartDir)
			} else {
				chrt, err = template.Load(templateType)
			}
			if err != nil {
				return fmt.Errorf("miramar deploy: %w", err)
			}

			overrides := map[string]interface{}{}
			if valuesFile != "" {
				fileVals, err := chartutil.ReadValuesFile(valuesFile)
				if err != nil {
					return fmt.Errorf("miramar deploy: reading %s: %w", valuesFile, err)
				}
				overrides = fileVals
			}
			if model != "" {
				setNestedString(overrides, model, "model", "id")
			}
			if servedModelName != "" {
				setNestedString(overrides, servedModelName, "model", "servedName")
			}

			client, err := deployer.NewClient(deployer.ClientOptions{
				KubeConfig: kubeconfigFlag,
				Namespace:  namespaceFlag,
			})
			if err != nil {
				return fmt.Errorf("miramar deploy: %w", err)
			}

			rel, err := client.Deploy(cmd.Context(), chrt, overrides, deployer.DeployOptions{
				ReleaseName: releaseNameFlag,
				Wait:        true,
				Timeout:     waitTimeout,
			})
			if err != nil {
				return fmt.Errorf("miramar deploy: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, styleOK(fmt.Sprintf("Deployed release %q (revision %d) in namespace %q", rel.Name, rel.Version, rel.Namespace)))
			fmt.Fprintln(out, "Run `miramar validate` to confirm the endpoint is serving requests.")
			return nil
		},
	}

	cmd.Flags().StringVar(&templateType, "type", "serving-vllm", "Built-in template type to deploy")
	cmd.Flags().StringVar(&chartDir, "chart-dir", "", "Deploy from a local chart directory instead of a built-in template (see: miramar new)")
	cmd.Flags().StringVar(&valuesFile, "values", "", "Additional Helm values file layered on top of the chart defaults")
	cmd.Flags().StringVar(&model, "model", "", "Override the Hugging Face model id")
	cmd.Flags().StringVar(&servedModelName, "served-model-name", "", "Override the served model name")
	cmd.Flags().DurationVar(&waitTimeout, "wait-timeout", 15*time.Minute, "How long to wait for the deployment to become ready")

	return cmd
}

// setNestedString sets value at the given dotted path within m, creating
// intermediate maps as needed and preserving any keys already present.
func setNestedString(m map[string]interface{}, value string, path ...string) {
	for i, key := range path {
		if i == len(path)-1 {
			m[key] = value
			return
		}
		next, ok := m[key].(map[string]interface{})
		if !ok {
			next = map[string]interface{}{}
			m[key] = next
		}
		m = next
	}
}
