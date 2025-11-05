package hclbuilder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Kong/shared-speakeasy/hclbuilder"
)

func TestNew_EmptyBuilder(t *testing.T) {
	builder := hclbuilder.New()
	require.NotNil(t, builder)

	result := builder.Build()
	require.Empty(t, result, "empty builder should produce empty HCL")
}

// updateGoldenFiles updates golden files when UPDATE_GOLDEN_FILES=1 is set
func updateGoldenFiles(t *testing.T, goldenFile string, actual string) bool {
	t.Helper()
	if os.Getenv("UPDATE_GOLDEN_FILES") == "1" {
		err := os.WriteFile(goldenFile, []byte(actual), 0o600)
		require.NoError(t, err, "updating golden file")
		t.Logf("updated golden file: %s", goldenFile)
		return true
	}
	return false
}

// assertGoldenFile compares actual output with golden file
func assertGoldenFile(t *testing.T, goldenFile string, actual string) {
	t.Helper()
	if updateGoldenFiles(t, goldenFile, actual) {
		return
	}

	expected, err := os.ReadFile(goldenFile)
	require.NoError(t, err, "reading golden file")
	require.Equal(t, string(expected), actual)
}

func TestSetAttribute_Simple(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetAttribute("resource.kong-mesh_mesh.default.name", "mesh-1")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "simple_attribute.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestSetAttribute_Multiple(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetAttribute("resource.kong-mesh_mesh.default.name", "mesh-1")
	builder.SetAttribute("resource.kong-mesh_mesh.default.type", "Mesh")
	builder.SetAttribute("resource.kong-mesh_mesh_traffic_permission.allow_all.mesh", "kong-mesh_mesh.default.name")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "multiple_attributes.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestSetBlock_Simple(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetBlock("resource.kong-mesh_mesh.default", map[string]any{
		"name":                           "mesh-1",
		"type":                           "Mesh",
		"skip_creating_initial_policies": []string{"*"},
	})

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "simple_block.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestSetBlock_Nested(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetBlock("resource.kong-mesh_mesh_traffic_permission.allow_all", map[string]any{
		"name": "allow-all",
		"type": "MeshTrafficPermission",
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

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "nested_block.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestRemoveAttribute(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetAttribute("resource.kong-mesh_mesh.default.name", "mesh-1")
	builder.SetAttribute("resource.kong-mesh_mesh.default.type", "Mesh")
	builder.RemoveAttribute("resource.kong-mesh_mesh.default.type")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "removed_attribute.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestRemoveBlock(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetBlock("resource.kong-mesh_mesh.default", map[string]any{
		"name": "mesh-1",
	})
	builder.SetBlock("resource.kong-mesh_mesh_traffic_permission.old_policy", map[string]any{
		"name": "old-policy",
	})
	builder.RemoveBlock("resource.kong-mesh_mesh_traffic_permission.old_policy")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "removed_block.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestFromFile(t *testing.T) {
	// Create a test file
	testFile := filepath.Join("testdata", "input.tf")
	builder, err := hclbuilder.FromFile(testFile)
	require.NoError(t, err)
	require.NotNil(t, builder)

	result := builder.Build()
	require.NotEmpty(t, result)
}

func TestFromFile_ModifyLoaded(t *testing.T) {
	testFile := filepath.Join("testdata", "input.tf")
	builder, err := hclbuilder.FromFile(testFile)
	require.NoError(t, err)

	// Modify the loaded configuration
	builder.SetAttribute("resource.kong-mesh_mesh.default.skip_creating_initial_policies", []string{"*"})
	builder.SetAttribute("resource.kong-mesh_mesh_traffic_permission.existing.type", "MeshTrafficPermission")
	builder.SetBlock("resource.kong-mesh_mesh_traffic_permission.new_policy", map[string]any{
		"name": "new-policy",
	})

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "modified_loaded.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestWriteFile(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetAttribute("variable.test.default", "value")

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.tf")

	err := builder.WriteFile(outputFile)
	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	require.Contains(t, string(content), "variable")
	require.Contains(t, string(content), "test")
}

func TestFromString(t *testing.T) {
	hclContent := `
resource "kong-mesh_mesh" "default" {
  name = "mesh-1"
  type = "Mesh"
}
`
	builder, err := hclbuilder.FromString(hclContent)
	require.NoError(t, err)
	require.NotNil(t, builder)

	result := builder.Build()
	require.Contains(t, result, "kong-mesh_mesh")
	require.Contains(t, result, "mesh-1")
}

func TestFromString_Invalid(t *testing.T) {
	invalidHCL := `resource "foo" {`
	_, err := hclbuilder.FromString(invalidHCL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing HCL")
}
