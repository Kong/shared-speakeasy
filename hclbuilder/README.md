# hclbuilder

Generic HCL builder for programmatically creating and modifying HCL configurations.

## Overview

`hclbuilder` provides a fluent API for building and modifying HCL (HashiCorp Configuration Language) files. Unlike the template-based `tfbuilder` package, this uses the official `hashicorp/hcl/v2/hclwrite` library for type-safe HCL manipulation.

## Installation

```bash
go get github.com/Kong/shared-speakeasy/hclbuilder
```

## Usage

### Create from scratch

```go
import "github.com/Kong/shared-speakeasy/hclbuilder"

builder := hclbuilder.New()
builder.SetBlock("resource.kong-mesh_mesh.default", map[string]any{
    "type": "Mesh",
    "name": "mesh-1",
    "skip_creating_initial_policies": []string{"*"},
})
builder.SetBlock("resource.kong-mesh_mesh_traffic_permission.allow_all", map[string]any{
    "type": "MeshTrafficPermission",
    "name": "allow-all",
    "mesh": "kong-mesh_mesh.default.name",
    "spec": map[string]any{
        "from": []any{
            map[string]any{
                "target_ref": map[string]any{
                    "kind": "Mesh",
                },
            },
        },
    },
})

hcl := builder.Build()
```

### Load from file

```go
builder, err := hclbuilder.FromFile("main.tf")
if err != nil {
    log.Fatal(err)
}

// Modify the loaded configuration
builder.SetAttribute("resource.kong-mesh_mesh.default.name", "mesh-2")
builder.RemoveBlock("resource.kong-mesh_mesh_traffic_permission.old_policy")

// Write back to file
err = builder.WriteFile("main.tf")
```

### Parse from string

```go
hclBlock := `
resource "kong-mesh_mesh" "default" {
  name = "mesh-1"
  type = "Mesh"
}
`
builder, err := hclbuilder.FromString(hclBlock)
if err != nil {
    log.Fatal(err)
}
// Use builder as needed
```

### Set attributes

```go
// Path format: "block_type.block_label.attribute_name"
builder.SetAttribute("resource.kong-mesh_mesh.default.name", "mesh-1")
builder.SetAttribute("resource.kong-mesh_mesh.default.type", "Mesh")
builder.SetAttribute("resource.kong-mesh_mesh_traffic_permission.allow_all.mesh", "kong-mesh_mesh.default.name")
```

### Create blocks

```go
// Simple block
builder.SetBlock("resource.kong-mesh_mesh.default", map[string]any{
    "type": "Mesh",
    "name": "mesh-1",
    "skip_creating_initial_policies": []string{"*"},
})

// Nested blocks (maps are treated as nested blocks)
builder.SetBlock("resource.kong-mesh_mesh_traffic_permission.allow_all", map[string]any{
    "type": "MeshTrafficPermission",
    "name": "allow-all",
    "mesh": "kong-mesh_mesh.default.name",
    "spec": map[string]any{
        "from": []any{
            map[string]any{
                "target_ref": map[string]any{
                    "kind": "Mesh",
                },
            },
        },
    },
})
```

### Remove attributes and blocks

```go
builder.RemoveAttribute("resource.kong-mesh_mesh.default.skip_creating_initial_policies")
builder.RemoveBlock("resource.kong-mesh_mesh_traffic_permission.old_policy")
```

## API

### Constructor Functions

- `New() *Builder` - Create empty builder
- `FromFile(path string) (*Builder, error)` - Load from HCL file
- `FromString(content string) (*Builder, error)` - Parse HCL from string

### Methods

- `Build() string` - Generate HCL string
- `WriteFile(path string) error` - Write to file
- `SetAttribute(path string, value any)` - Set attribute value
- `SetBlock(path string, attributes map[string]any)` - Create/replace block
- `RemoveAttribute(path string)` - Remove attribute
- `RemoveBlock(path string)` - Remove block

### Path Format

Paths use dot notation:
- Attributes: `"block_type.block_label.attribute_name"`
- Blocks: `"block_type.block_label1.block_label2"`

Examples:
- `"variable.mesh_name.default"` → `variable "mesh_name" { default = ... }`
- `"resource.kong-mesh_mesh.default"` → `resource "kong-mesh_mesh" "default" { ... }`

## Testing

```bash
# Run tests
go test ./...

# Update golden files
UPDATE_GOLDEN_FILES=1 go test ./...
```

## Test Helpers

The package includes test helpers for Terraform provider testing:

### TestBuilder

Provider-specific builder for generating test configurations:

```go
import "github.com/Kong/shared-speakeasy/hclbuilder"

// Create a test builder
builder := hclbuilder.NewTestBuilder(hclbuilder.KongMesh)

// Add resources
builder.AddMesh("mesh-1", "default", map[string]any{
    "skip_creating_initial_policies": []string{"*"},
})

// Generate HCL
config := builder.Build()
```

### Test Case Functions

Reusable test cases for common scenarios:

- `CreateMeshAndModifyFields` - Tests mesh creation and field modifications
- `CreatePolicyAndModifyFields` - Tests policy creation and field modifications
- `NotImportedResourceShouldError` - Tests error handling for non-imported resources

See `test_cases.go` for details.

## Differences from tfbuilder

- **Generic**: Not tied to specific Terraform providers
- **Type-safe**: Uses `hclwrite` library instead of templates
- **File I/O**: Can read and modify existing HCL files
- **Simpler API**: Unified interface for all HCL manipulation
