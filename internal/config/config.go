package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DeploymentConfig mirrors the relevant parts of a Kubernetes
// apps/v1 Deployment manifest.
type DeploymentConfig struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   Metadata       `yaml:"metadata"`
	Spec       DeploymentSpec `yaml:"spec"`
}

type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type DeploymentSpec struct {
	Replicas int         `yaml:"replicas"`
	Template PodTemplate `yaml:"template"`
}

type PodTemplate struct {
	Spec PodSpec `yaml:"spec"`
}

type PodSpec struct {
	Containers []Container `yaml:"containers"`
}

type Container struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

// LoadConfig reads a YAML file from the given path and parses it
// into a DeploymentConfig.
func LoadConfig(path string) (*DeploymentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var cfg DeploymentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	return &cfg, nil
}

// Validate checks that required fields are present and sane.
func Validate(cfg *DeploymentConfig) error {
	if cfg.Kind != "Deployment" {
		return fmt.Errorf("kind must be Deployment, got %q", cfg.Kind)
	}
	if cfg.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if cfg.Spec.Replicas <= 0 {
		return fmt.Errorf("spec.replicas must be greater than 0")
	}

	containers := cfg.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return fmt.Errorf("spec.template.spec.containers must have at least one container")
	}
	for i, c := range containers {
		if c.Name == "" {
			return fmt.Errorf("spec.template.spec.containers[%d].name is required", i)
		}
		if c.Image == "" {
			return fmt.Errorf("spec.template.spec.containers[%d].image is required", i)
		}
	}

	return nil
}
