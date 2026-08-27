# configvalidator

A small Go CLI that validates the shape of a Kubernetes `apps/v1` Deployment manifest (YAML).

It checks that:
- `kind` is `Deployment`
- `metadata.name` is set
- `spec.replicas` is greater than `0`
- `spec.template.spec.containers` has at least one entry
- every container has a `name` and an `image`

## Usage

```bash
go run main.go path/to/deployment.yaml
```

On success:

```
config is valid: name=payment-service namespace=default replicas=3 containers=1
```

On failure, it prints the reason and exits with a non-zero status:

```
config is invalid: spec.replicas must be greater than 0
```

## Run tests

```bash
go test ./...
```

## Project structure

```
configvalidator/
  main.go                     # CLI entry point
  internal/
    config/
      config.go                # Deployment struct, LoadConfig, Validate
      config_test.go           # table-driven tests
```

## Notes

- Currently supports the `Deployment` kind only. Other kinds (`Service`, `ConfigMap`, etc.) are not yet parsed.
- Only a subset of Deployment fields are modeled (`metadata.name`, `metadata.namespace`, `spec.replicas`, `spec.template.spec.containers[].name/image`). Unrecognized YAML fields are ignored, not rejected.