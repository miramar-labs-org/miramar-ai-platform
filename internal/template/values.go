package template

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PatchValues rewrites specific fields of destDir/values.yaml, produced by an
// earlier Copy, in place. It edits the parsed yaml.Node tree rather than
// round-tripping through a generic map so that comments and key ordering in
// the file survive. Empty arguments are left untouched — a no-op if both are
// empty.
func PatchValues(destDir, modelID, servedModelName string) error {
	if modelID == "" && servedModelName == "" {
		return nil
	}

	path := filepath.Join(destDir, "values.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return fmt.Errorf("%s is empty", path)
	}
	root := doc.Content[0]

	model := mappingValue(root, "model")
	if model == nil {
		return fmt.Errorf("%s has no top-level \"model\" key", path)
	}

	if modelID != "" {
		if err := setMappingValue(model, "id", modelID); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	if servedModelName != "" {
		if err := setMappingValue(model, "servedName", servedModelName); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(&doc)
}

// mappingValue returns the value node for key within a YAML mapping node, or
// nil if node isn't a mapping or key isn't present.
func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

// setMappingValue sets the scalar value for key within a YAML mapping node.
func setMappingValue(node *yaml.Node, key, value string) error {
	v := mappingValue(node, key)
	if v == nil {
		return fmt.Errorf("no %q key found", key)
	}
	if v.Kind != yaml.ScalarNode {
		return fmt.Errorf("%q is not a scalar value", key)
	}
	v.Value = value
	return nil
}
