package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Validatable is satisfied by any manifest type that can validate
// itself and describe itself in one line.
type Validatable interface {
	Validate() error
	Summary() string
}

// Deployment mirrors the relevant parts of a Kubernetes apps/v1 Deployment.
type Deployment struct {
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

// Validate checks that required Deployment fields are present and sane.
func (d *Deployment) Validate() error {
	if d.Kind != "Deployment" {
		return fmt.Errorf("kind must be Deployment, got %q", d.Kind)
	}
	if d.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if d.Spec.Replicas <= 0 {
		return fmt.Errorf("spec.replicas must be greater than 0")
	}

	containers := d.Spec.Template.Spec.Containers
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

// Service mirrors the relevant parts of a Kubernetes v1 Service.
type Service struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       ServiceSpec `yaml:"spec"`
}

type ServiceSpec struct {
	Selector map[string]string `yaml:"selector"`
	Ports    []ServicePort     `yaml:"ports"`
}

type ServicePort struct {
	Port       int `yaml:"port"`
	TargetPort int `yaml:"targetPort"`
}

func (s *Service) Validate() error {
	if s.Kind != "Service" {
		return fmt.Errorf("kind must be Service, got %q", s.Kind)
	}
	if s.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if len(s.Spec.Selector) == 0 {
		return fmt.Errorf("spec.selector must have at least one entry")
	}
	if len(s.Spec.Ports) == 0 {
		return fmt.Errorf("spec.ports must have at least one port")
	}
	for i, p := range s.Spec.Ports {
		if p.Port <= 0 {
			return fmt.Errorf("spec.ports[%d].port must be greater than 0", i)
		}
	}
	return nil
}

func (s *Service) Summary() string {
	return fmt.Sprintf("Service name=%s namespace=%s ports=%d",
		s.Metadata.Name, s.Metadata.Namespace, len(s.Spec.Ports))
}

// ConfigMap mirrors the relevant parts of a Kubernetes v1 ConfigMap.
type ConfigMap struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   Metadata          `yaml:"metadata"`
	Data       map[string]string `yaml:"data"`
}

func (c *ConfigMap) Validate() error {
	if c.Kind != "ConfigMap" {
		return fmt.Errorf("kind must be ConfigMap, got %q", c.Kind)
	}
	if c.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if len(c.Data) == 0 {
		return fmt.Errorf("data must have at least one entry")
	}
	return nil
}

func (c *ConfigMap) Summary() string {
	return fmt.Sprintf("ConfigMap name=%s namespace=%s keys=%d",
		c.Metadata.Name, c.Metadata.Namespace, len(c.Data))
}

// StatefulSet mirrors the relevant parts of a Kubernetes apps/v1 StatefulSet.
type StatefulSet struct {
	APIVersion string          `yaml:"apiVersion"`
	Kind       string          `yaml:"kind"`
	Metadata   Metadata        `yaml:"metadata"`
	Spec       StatefulSetSpec `yaml:"spec"`
}

type StatefulSetSpec struct {
	ServiceName string      `yaml:"serviceName"`
	Replicas    int         `yaml:"replicas"`
	Template    PodTemplate `yaml:"template"`
}

func (s *StatefulSet) Validate() error {
	if s.Kind != "StatefulSet" {
		return fmt.Errorf("kind must be StatefulSet, got %q", s.Kind)
	}
	if s.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if s.Spec.ServiceName == "" {
		return fmt.Errorf("spec.serviceName is required")
	}
	if s.Spec.Replicas <= 0 {
		return fmt.Errorf("spec.replicas must be greater than 0")
	}

	containers := s.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return fmt.Errorf("spec.template.spec.containers must have at least one container")
	}
	for i, c := range containers {
		if c.Image == "" {
			return fmt.Errorf("spec.template.spec.containers[%d].image is required", i)
		}
	}

	return nil
}

func (s *StatefulSet) Summary() string {
	return fmt.Sprintf("StatefulSet name=%s namespace=%s replicas=%d",
		s.Metadata.Name, s.Metadata.Namespace, s.Spec.Replicas)
}

// Summary returns a one-line human-readable description.
func (d *Deployment) Summary() string {
	return fmt.Sprintf("Deployment name=%s namespace=%s replicas=%d containers=%d",
		d.Metadata.Name, d.Metadata.Namespace, d.Spec.Replicas, len(d.Spec.Template.Spec.Containers))
}

// kindPeek is used only to read the `kind` field before deciding
// which concrete type to fully parse into.
type kindPeek struct {
	Kind string `yaml:"kind"`
}

// LoadManifest reads a YAML file, identifies its kind, and parses it
// into the matching concrete type, returned as a Validatable.
// Returns an error if the file can't be read/parsed, or if the kind
// is not one we support.
func LoadManifest(path string) (Validatable, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var peek kindPeek
	if err := yaml.Unmarshal(data, &peek); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	var manifest Validatable

	switch peek.Kind {
	case "Deployment":
		manifest = &Deployment{}
	case "Service":
		manifest = &Service{}
	case "ConfigMap":
		manifest = &ConfigMap{}
	case "StatefulSet":
		manifest = &StatefulSet{}
	default:
		return nil, fmt.Errorf("unsupported kind: %q", peek.Kind)
	}

	if err := yaml.Unmarshal(data, manifest); err != nil {
		return nil, fmt.Errorf("parsing yaml as %s: %w", peek.Kind, err)
	}

	return manifest, nil
}
