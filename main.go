package main

import (
	"fmt"
	"os"

	"github.com/blangaa/configvalidator/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: configvalidator <path-to-config.yaml>")
		os.Exit(1)
	}

	path := os.Args[1]

	cfg, err := config.LoadConfig(path)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := config.Validate(cfg); err != nil {
		fmt.Printf("config is invalid: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("config is valid: name=%s namespace=%s replicas=%d containers=%d\n",
		cfg.Metadata.Name, cfg.Metadata.Namespace, cfg.Spec.Replicas, len(cfg.Spec.Template.Spec.Containers))
}