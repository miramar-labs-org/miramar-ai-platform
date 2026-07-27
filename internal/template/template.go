// Package template implements the deployment-template factory: resolving a
// named template type to a loadable Helm chart (from the binary's embedded
// templates or an on-disk directory), and copying a template out to a local
// directory for a user to own and customize before deploying from it.
package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/miramar-labs-org/miramar-ai-platform/templates"
)

// Load builds a chart.Chart from the embedded template named templateType.
func Load(templateType string) (*chart.Chart, error) {
	if !slices.Contains(templates.Available, templateType) {
		return nil, fmt.Errorf("unknown template type %q, available: %s", templateType, strings.Join(templates.Available, ", "))
	}

	var files []*loader.BufferedFile
	root := templateType
	err := fs.WalkDir(templates.FS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(templates.FS, path)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(path, root), "/")
		files = append(files, &loader.BufferedFile{Name: rel, Data: data})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading embedded template %q: %w", templateType, err)
	}

	return loader.LoadFiles(files)
}

// LoadDir loads a chart from an on-disk directory, e.g. one produced by Copy
// and then customized by hand.
func LoadDir(path string) (*chart.Chart, error) {
	return loader.LoadDir(path)
}

// Copy writes the embedded template named templateType to destDir, preserving
// its directory structure. It refuses to write into a destDir that already
// exists and is non-empty unless force is set.
func Copy(templateType, destDir string, force bool) error {
	if !slices.Contains(templates.Available, templateType) {
		return fmt.Errorf("unknown template type %q, available: %s", templateType, strings.Join(templates.Available, ", "))
	}

	if !force {
		entries, err := os.ReadDir(destDir)
		if err == nil && len(entries) > 0 {
			return fmt.Errorf("destination %q already exists and is not empty (use --force to overwrite)", destDir)
		}
	}

	root := templateType
	return fs.WalkDir(templates.FS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(path, root), "/")
		if rel == "" {
			return os.MkdirAll(destDir, 0o755)
		}
		target := filepath.Join(destDir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(templates.FS, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
