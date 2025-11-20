package hclbuilder_test

import (
	"testing"

	"github.com/Kong/shared-speakeasy/hclbuilder"
	"github.com/stretchr/testify/require"
)

func TestBuilder_EmbedAndMutateExample(t *testing.T) {
	// Create a main file with a provider
	mainFile, err := hclbuilder.FromString(`
provider "kong-mesh" {
  server_url = "http://localhost:5681"
}
`)
	require.NoError(t, err)

	// Create a mesh resource
	mesh, err := hclbuilder.FromString(`
resource "kong-mesh_mesh" "default" {
  type = "Mesh"
  name = "default"
}
`)
	require.NoError(t, err)

	// Add mesh to main file
	mainFile.Add(mesh)

	output1 := mainFile.Build()
	t.Logf("=== Initial Build ===\n%s", output1)
	require.Contains(t, output1, `provider "kong-mesh"`)
	require.Contains(t, output1, `resource "kong-mesh_mesh" "default"`)

	// Modify the mesh by adding an attribute
	mesh.AddAttribute("skip_creating_initial_policies", []string{"*"})

	output2 := mainFile.Build()
	t.Logf("=== After adding skip_creating_initial_policies ===\n%s", output2)
	require.Contains(t, output2, "skip_creating_initial_policies")

	// Add a nested attribute
	mesh.AddAttribute("routing.default_forbid_mesh_external_service_access", true)

	output3 := mainFile.Build()
	t.Logf("=== After adding routing attribute ===\n%s", output3)
	require.Contains(t, output3, "routing")
	require.Contains(t, output3, "default_forbid_mesh_external_service_access")

	// Remove an attribute
	mesh.RemoveAttribute("skip_creating_initial_policies")

	output4 := mainFile.Build()
	t.Logf("=== After removing skip_creating_initial_policies ===\n%s", output4)
	require.NotContains(t, output4, "skip_creating_initial_policies")
	require.Contains(t, output4, "routing") // routing should still be there
}
