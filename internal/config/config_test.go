package config

import "testing"

func validDeployment() DeploymentConfig {
	return DeploymentConfig{
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

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(cfg *DeploymentConfig)
		wantErr bool
	}{
		{
			name:    "valid config",
			modify:  func(cfg *DeploymentConfig) {},
			wantErr: false,
		},
		{
			name:    "wrong kind",
			modify:  func(cfg *DeploymentConfig) { cfg.Kind = "Service" },
			wantErr: true,
		},
		{
			name:    "missing name",
			modify:  func(cfg *DeploymentConfig) { cfg.Metadata.Name = "" },
			wantErr: true,
		},
		{
			name:    "zero replicas",
			modify:  func(cfg *DeploymentConfig) { cfg.Spec.Replicas = 0 },
			wantErr: true,
		},
		{
			name:    "no containers",
			modify:  func(cfg *DeploymentConfig) { cfg.Spec.Template.Spec.Containers = nil },
			wantErr: true,
		},
		{
			name: "container missing image",
			modify: func(cfg *DeploymentConfig) {
				cfg.Spec.Template.Spec.Containers[0].Image = ""
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validDeployment()
			tt.modify(&cfg)

			err := Validate(&cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
