package hclbuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder_EmbedAndMutate(t *testing.T) {
	// Create a main file with a provider
	mainFile, err := FromString(`
provider "kong-mesh" {
  server_url = "http://localhost:5681"
}
`)
	require.NoError(t, err)

	// Create a mesh resource
	mesh, err := FromString(`
resource "kong-mesh_mesh" "default" {
  type = "Mesh"
  name = "default"
}
`)
	require.NoError(t, err)

	// Add mesh to main file
	mainFile.Add(mesh)

	output1 := mainFile.Build()
	require.Contains(t, output1, `provider "kong-mesh"`)
	require.Contains(t, output1, `resource "kong-mesh_mesh" "default"`)
	require.NotContains(t, output1, "skip_creating_initial_policies")

	// Now modify the mesh by adding an attribute
	mesh.AddAttribute("skip_creating_initial_policies", []string{"*"})

	// Build main file again - should reflect the change
	output2 := mainFile.Build()
	require.Contains(t, output2, `provider "kong-mesh"`)
	require.Contains(t, output2, `resource "kong-mesh_mesh" "default"`)
	require.Contains(t, output2, "skip_creating_initial_policies")
}

func TestBuilder_AddAttributeNested(t *testing.T) {
	mesh, err := FromString(`
resource "kong-mesh_mesh" "default" {
  type = "Mesh"
  name = "default"
}
`)
	require.NoError(t, err)

	mesh.AddAttribute("routing.default_forbid_mesh_external_service_access", true)

	output := mesh.Build()
	require.Contains(t, output, "routing")
	require.Contains(t, output, "default_forbid_mesh_external_service_access")
	require.Contains(t, output, "= true")
}

func TestBuilder_RemoveAttributeSimple(t *testing.T) {
	mesh, err := FromString(`
resource "kong-mesh_mesh" "default" {
  type = "Mesh"
  name = "default"
  skip_creating_initial_policies = ["*"]
}
`)
	require.NoError(t, err)

	mesh.RemoveAttribute("skip_creating_initial_policies")

	output := mesh.Build()
	t.Log(output)
	require.NotContains(t, output, "skip_creating_initial_policies")
	// Should still have the resource
	require.Contains(t, output, `resource "kong-mesh_mesh" "default"`)
}

func TestBuilder_AddAttributeFromHCLString(t *testing.T) {
	mesh, err := FromString(`
resource "kong-mesh_mesh" "default" {
  type = "Mesh"
  name = "default"
}
`)
	require.NoError(t, err)

	// Test simple list from HCL string
	mesh.AddAttribute("skip_creating_initial_policies", `["*"]`)

	output := mesh.Build()
	t.Log(output)
	require.Contains(t, output, `skip_creating_initial_policies = ["*"]`)

	// Test nested object from HCL string
	mesh.AddAttribute("constraints.dataplane_proxy.requirements", `[{ tags = { key = "a" } }]`)

	output = mesh.Build()
	t.Log(output)
	require.Contains(t, output, "constraints")
	require.Contains(t, output, "dataplane_proxy")
	require.Contains(t, output, "requirements")
	require.Contains(t, output, "tags")
	require.Contains(t, output, `key = "a"`)
}

func TestBuilder_AddAttributeFromHCLStringComplex(t *testing.T) {
	mesh, err := FromString(`
resource "kong-mesh_mesh" "default" {
  type = "Mesh"
  name = "default"
}
`)
	require.NoError(t, err)

	// Test boolean
	mesh.AddAttribute("routing.default_forbid_mesh_external_service_access", "true")

	output := mesh.Build()
	t.Log(output)
	require.Contains(t, output, "routing")
	require.Contains(t, output, "default_forbid_mesh_external_service_access = true")

	// Test empty list
	mesh.AddAttribute("constraints.dataplane_proxy.restrictions", "[]")

	output = mesh.Build()
	t.Log(output)
	require.Contains(t, output, "restrictions = []")
}
