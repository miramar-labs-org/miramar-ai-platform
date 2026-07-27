package main

import "github.com/spf13/cobra"

func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install platform-side components (namespace, secrets, etc.)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return notImplemented("install")
		},
	}
}
