# configvalidator

A Go CLI that validates Kubernetes manifests (YAML) — either a single file or an entire directory tree.

## Supported kinds

- `Deployment` (apps/v1)
- `StatefulSet` (apps/v1)
- `Service` (v1)
- `ConfigMap` (v1)

Any other `kind` is reported as "could not process" rather than treated as invalid — the tool doesn't have rules for it yet, but the manifest itself might be perfectly fine.

## What's validated

**Deployment / StatefulSet**
- `metadata.name` is set
- `spec.replicas` is greater than `0`
- `spec.template.spec.containers` has at least one entry, each with a `name` and `image`
- (StatefulSet only) `spec.serviceName` is set

**Service**
- `metadata.name` is set
- `spec.selector` has at least one entry
- `spec.ports` has at least one entry, each with `port` greater than `0`

**ConfigMap**
- `metadata.name` is set
- `data` has at least one entry

## Usage

Single file:

```bash
go run main.go path/to/deployment.yaml
```

Directory (recursively finds all `.yaml`/`.yml` files):

```bash
go run main.go path/to/manifests/
```

Example output:

```
✅ manifests/configmap.yaml: ConfigMap name=app-config namespace=default keys=2
✅ manifests/deployment.yaml: Deployment name=payment-service namespace=default replicas=3 containers=1
❌ manifests/broken.yaml: invalid (spec.replicas must be greater than 0)
⚠️  manifests/unsupported.yaml: could not process (unsupported kind: "Ingress")

Summary: 2 valid, 1 invalid, 1 skipped (4 total)
```

Exit code is non-zero if any manifest is invalid. Skipped (unsupported kind / parse error) manifests don't affect the exit code.

## Run tests

```bash
go test ./...
```

## Project structure

```
configvalidator/
  main.go                     # CLI entry point, directory walking, output formatting
  internal/
    config/
      config.go                # manifest types, Validatable interface, LoadManifest dispatcher
      config_test.go           # table-driven tests per manifest kind
```

## Design notes

- All manifest types implement a small `Validatable` interface (`Validate() error`, `Summary() string`), which is why `main.go` doesn't need to know or care which concrete kind it's holding.
- `LoadManifest` does a two-pass YAML parse: first to read just the `kind` field, then to fully parse into the matching concrete type.
- Only a subset of fields are modeled per kind. Unrecognized YAML fields are ignored, not rejected.