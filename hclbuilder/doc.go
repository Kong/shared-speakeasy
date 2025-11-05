// Package hclbuilder provides a fluent API for building and modifying HCL configurations.
//
// This package uses the official hashicorp/hcl/v2/hclwrite library for type-safe
// HCL manipulation, providing a simpler interface for common operations.
//
// Example usage:
//
//	// Create from scratch
//	builder := hclbuilder.New()
//	builder.SetAttribute("variable.region.default", "us-west-2")
//	builder.SetBlock("resource.aws_instance.web", map[string]any{
//	    "ami": "ami-123456",
//	    "instance_type": "t2.micro",
//	})
//	hcl := builder.Build()
//
//	// Load and modify existing file
//	builder, err := hclbuilder.FromFile("main.tf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	builder.RemoveBlock("resource.aws_instance.old")
//	err = builder.WriteFile("main.tf")
package hclbuilder
