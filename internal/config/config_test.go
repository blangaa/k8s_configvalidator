package config

import (
	"os"
	"strings"
	"testing"
)

func validDeployment() Deployment {
	return Deployment{
		Kind: "Deployment",
		Metadata: Metadata{
			Name:      "payment-service",
			Namespace: "default",
		},
		Spec: DeploymentSpec{
			Replicas: 3,
			Template: PodTemplate{
				Spec: PodSpec{
					Containers: []Container{
						{Name: "payment-service", Image: "myregistry/payment-service:1.4.2"},
					},
				},
			},
		},
	}
}

func TestDeploymentValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(d *Deployment)
		wantErr bool
	}{
		{"valid", func(d *Deployment) {}, false},
		{"wrong kind", func(d *Deployment) { d.Kind = "Service" }, true},
		{"missing name", func(d *Deployment) { d.Metadata.Name = "" }, true},
		{"zero replicas", func(d *Deployment) { d.Spec.Replicas = 0 }, true},
		{"no containers", func(d *Deployment) { d.Spec.Template.Spec.Containers = nil }, true},
		{"container missing image", func(d *Deployment) { d.Spec.Template.Spec.Containers[0].Image = "" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := validDeployment()
			tt.modify(&d)
			err := d.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validService() Service {
	return Service{
		Kind:     "Service",
		Metadata: Metadata{Name: "payment-service", Namespace: "default"},
		Spec: ServiceSpec{
			Selector: map[string]string{"app": "payment-service"},
			Ports:    []ServicePort{{Port: 80, TargetPort: 8080}},
		},
	}
}

func TestServiceValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(s *Service)
		wantErr bool
	}{
		{"valid", func(s *Service) {}, false},
		{"wrong kind", func(s *Service) { s.Kind = "Deployment" }, true},
		{"missing name", func(s *Service) { s.Metadata.Name = "" }, true},
		{"no selector", func(s *Service) { s.Spec.Selector = nil }, true},
		{"no ports", func(s *Service) { s.Spec.Ports = nil }, true},
		{"zero port", func(s *Service) { s.Spec.Ports[0].Port = 0 }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := validService()
			tt.modify(&s)
			err := s.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validConfigMap() ConfigMap {
	return ConfigMap{
		Kind:     "ConfigMap",
		Metadata: Metadata{Name: "app-config", Namespace: "default"},
		Data:     map[string]string{"LOG_LEVEL": "info"},
	}
}

func TestConfigMapValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(c *ConfigMap)
		wantErr bool
	}{
		{"valid", func(c *ConfigMap) {}, false},
		{"wrong kind", func(c *ConfigMap) { c.Kind = "Secret" }, true},
		{"missing name", func(c *ConfigMap) { c.Metadata.Name = "" }, true},
		{"no data", func(c *ConfigMap) { c.Data = nil }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfigMap()
			tt.modify(&c)
			err := c.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validStatefulSet() StatefulSet {
	return StatefulSet{
		Kind:     "StatefulSet",
		Metadata: Metadata{Name: "cassandra", Namespace: "default"},
		Spec: StatefulSetSpec{
			ServiceName: "cassandra-headless",
			Replicas:    3,
			Template: PodTemplate{
				Spec: PodSpec{
					Containers: []Container{
						{Name: "cassandra", Image: "cassandra:4.1"},
					},
				},
			},
		},
	}
}

func TestStatefulSetValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(s *StatefulSet)
		wantErr bool
	}{
		{"valid", func(s *StatefulSet) {}, false},
		{"wrong kind", func(s *StatefulSet) { s.Kind = "Deployment" }, true},
		{"missing name", func(s *StatefulSet) { s.Metadata.Name = "" }, true},
		{"missing service name", func(s *StatefulSet) { s.Spec.ServiceName = "" }, true},
		{"zero replicas", func(s *StatefulSet) { s.Spec.Replicas = 0 }, true},
		{"no containers", func(s *StatefulSet) { s.Spec.Template.Spec.Containers = nil }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := validStatefulSet()
			tt.modify(&s)
			err := s.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadManifest(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		wantErrText string
	}{
		{
			name: "valid deployment",
			yaml: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: payment-service
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: payment-service
          image: myregistry/payment-service:1.4.2
`,
			wantErr: false,
		},
		{
			name: "unsupported kind",
			yaml: `
apiVersion: v1
kind: Ingress
metadata:
  name: some-ingress
`,
			wantErr:     true,
			wantErrText: "unsupported kind",
		},
		{
			name:        "malformed yaml",
			yaml:        `not: [valid: yaml`,
			wantErr:     true,
			wantErrText: "parsing yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := dir + "/manifest.yaml"
			if err := os.WriteFile(path, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			manifest, err := LoadManifest(path)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrText) {
					t.Errorf("expected error to contain %q, got %q", tt.wantErrText, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if manifest == nil {
				t.Fatal("expected non-nil manifest")
			}
		})
	}
}

func TestLoadManifest_FileNotFound(t *testing.T) {
	_, err := LoadManifest("/nonexistent/path/manifest.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
