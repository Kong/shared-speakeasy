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
builder.SetAttribute("variable.name.default", "my-value")
builder.SetBlock("resource.aws_instance.web", map[string]interface{}{
    "ami": "ami-123456",
    "instance_type": "t2.micro",
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
builder.SetAttribute("variable.region.default", "us-west-2")
builder.RemoveBlock("resource.aws_instance.old")

// Write back to file
err = builder.WriteFile("main.tf")
```

### Set attributes

```go
// Path format: "block_type.block_label.attribute_name"
builder.SetAttribute("variable.name.default", "test-value")
builder.SetAttribute("variable.name.description", "A test variable")
builder.SetAttribute("variable.count.default", 5)
```

### Create blocks

```go
// Simple block
builder.SetBlock("resource.aws_instance.web", map[string]interface{}{
    "ami": "ami-123456",
    "instance_type": "t2.micro",
})

// Nested blocks (maps are treated as nested blocks)
builder.SetBlock("resource.aws_security_group.example", map[string]interface{}{
    "name": "example",
    "ingress": map[string]interface{}{
        "from_port": 80,
        "to_port": 80,
        "protocol": "tcp",
        "cidr_blocks": []string{"0.0.0.0/0"},
    },
})
```

### Remove attributes and blocks

```go
builder.RemoveAttribute("variable.name.sensitive")
builder.RemoveBlock("resource.aws_instance.old")
```

## API

### Constructor Functions

- `New() *Builder` - Create empty builder
- `FromFile(path string) (*Builder, error)` - Load from HCL file

### Methods

- `Build() string` - Generate HCL string
- `WriteFile(path string) error` - Write to file
- `SetAttribute(path string, value interface{})` - Set attribute value
- `SetBlock(path string, attributes map[string]interface{})` - Create/replace block
- `RemoveAttribute(path string)` - Remove attribute
- `RemoveBlock(path string)` - Remove block

### Path Format

Paths use dot notation:
- Attributes: `"block_type.block_label.attribute_name"`
- Blocks: `"block_type.block_label1.block_label2"`

Examples:
- `"variable.name.default"` → `variable "name" { default = ... }`
- `"resource.aws_instance.web"` → `resource "aws_instance" "web" { ... }`

## Testing

```bash
# Run tests
go test ./...

# Update golden files
UPDATE_GOLDEN_FILES=1 go test ./...
```

## Differences from tfbuilder

- **Generic**: Not tied to specific Terraform providers
- **Type-safe**: Uses `hclwrite` library instead of templates
- **File I/O**: Can read and modify existing HCL files
- **Simpler API**: Unified interface for all HCL manipulation
