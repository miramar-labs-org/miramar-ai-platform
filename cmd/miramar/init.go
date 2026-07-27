package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/miramar-labs-org/miramar-ai-platform/internal/template"
)

func newInitCmd() *cobra.Command {
	var (
		templateType    string
		dir             string
		model           string
		servedModelName string
		force           bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Copy a deployment template to a local directory for customization",
		Long: `init copies one of the deployment templates this binary ships with into a
local directory you own, optionally patching the model id and served name in
its values.yaml. Edit the copy further by hand, then run:

  miramar deploy --chart-dir <dir>

to deploy from the customized copy instead of the built-in default.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			destDir := dir
			if destDir == "" {
				destDir = templateType
			}

			if err := template.Copy(templateType, destDir, force); err != nil {
				return fmt.Errorf("miramar init: %w", err)
			}
			if err := template.PatchValues(destDir, model, servedModelName); err != nil {
				return fmt.Errorf("miramar init: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Copied template %q to %s\n", templateType, destDir)
			fmt.Fprintf(out, "Edit %s/values.yaml as needed, then run:\n  miramar deploy --chart-dir %s\n", destDir, destDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&templateType, "type", "serving-vllm", "Template type to copy")
	cmd.Flags().StringVar(&dir, "dir", "", "Destination directory (default: ./<type>)")
	cmd.Flags().StringVar(&model, "model", "", "Override the Hugging Face model id in the copied values.yaml")
	cmd.Flags().StringVar(&servedModelName, "served-model-name", "", "Override the served model name in the copied values.yaml")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite an existing non-empty destination directory")

	return cmd
}
