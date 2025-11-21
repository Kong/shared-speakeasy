// Package hclbuilder provides a fluent API for building and modifying HCL configurations.
//
// This package uses the official hashicorp/hcl/v2/hclwrite library for type-safe
// HCL manipulation, providing a simpler interface for common operations.
//
// Example usage:
//
//	// Create from scratch
//	builder := hclbuilder.New()
//	builder.SetBlock("resource.kong-mesh_mesh.default", map[string]any{
//	    "type": "Mesh",
//	    "name": "mesh-1",
//	    "skip_creating_initial_policies": []string{"*"},
//	})
//	builder.SetBlock("resource.kong-mesh_mesh_traffic_permission.allow_all", map[string]any{
//	    "type": "MeshTrafficPermission",
//	    "name": "allow-all",
//	    "mesh": "kong-mesh_mesh.default.name",
//	    "spec": map[string]any{
//	        "from": []any{
//	            map[string]any{
//	                "target_ref": map[string]any{
//	                    "kind": "Mesh",
//	                },
//	            },
//	        },
//	    },
//	})
//	hcl := builder.Build()
//
//	// Load and modify existing file
//	builder, err := hclbuilder.FromFile("main.tf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	builder.SetAttribute("resource.kong-mesh_mesh.default.skip_creating_initial_policies", []string{"*"})
//	err = builder.WriteFile("main.tf")
package hclbuilder
