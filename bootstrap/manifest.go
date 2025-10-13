package bootstrap

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Version string `yaml:"version"`
	BinDir  string `yaml:"binDir"`
	Tools   []Tool `yaml:"tools"`
}

type Tool struct {
	ID          string  `yaml:"id"`
	Description string  `yaml:"description"`
	Required    bool    `yaml:"required"`
	Install     Install `yaml:"install"`
}

type Install struct {
	Type        string            `yaml:"type"`
	Module      string            `yaml:"module,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	Command     string            `yaml:"command,omitempty"`
	URL         string            `yaml:"url,omitempty"`
	Source      string            `yaml:"source,omitempty"`
	BinName     string            `yaml:"binName,omitempty"`
	Destination string            `yaml:"destination,omitempty"`
	Checksum    map[string]string `yaml:"checksum,omitempty"`
}

func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ManifestError{Path: path, Err: err}
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, &ManifestError{Path: path, Err: fmt.Errorf("invalid YAML: %w", err)}
	}

	if err := validateManifest(&manifest); err != nil {
		return nil, &ManifestError{Path: path, Err: err}
	}

	return &manifest, nil
}

func validateManifest(m *Manifest) error {
	if m.Version == "" {
		return fmt.Errorf("missing required field: version")
	}

	if len(m.Tools) == 0 {
		return fmt.Errorf("no tools defined")
	}

	for i, tool := range m.Tools {
		if err := validateTool(&tool); err != nil {
			return fmt.Errorf("tool[%d] (%s): %w", i, tool.ID, err)
		}
	}

	return nil
}

func validateTool(t *Tool) error {
	if t.ID == "" {
		return fmt.Errorf("missing required field: id")
	}

	if t.Install.Type == "" {
		return fmt.Errorf("missing required field: install.type")
	}

	switch t.Install.Type {
	case "go":
		if t.Install.Module == "" {
			return fmt.Errorf("type 'go' requires 'module' field")
		}
		if t.Install.Version == "" {
			return fmt.Errorf("type 'go' requires 'version' field")
		}

	case "verify":
		if t.Install.Command == "" {
			return fmt.Errorf("type 'verify' requires 'command' field")
		}

	case "download":
		if t.Install.URL == "" {
			return fmt.Errorf("type 'download' requires 'url' field")
		}
		if t.Install.BinName == "" {
			return fmt.Errorf("type 'download' requires 'binName' field")
		}

	case "link":
		if t.Install.Source == "" {
			return fmt.Errorf("type 'link' requires 'source' field")
		}
		if t.Install.BinName == "" {
			return fmt.Errorf("type 'link' requires 'binName' field")
		}

	default:
		return fmt.Errorf("unsupported install type: %s", t.Install.Type)
	}

	return nil
}
