package hclbuilder_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Kong/shared-speakeasy/hclbuilder"
)

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

// Test New() - empty builder
func TestNew_EmptyBuilder(t *testing.T) {
	builder := hclbuilder.New()
	require.NotNil(t, builder)

	result := builder.Build()
	require.Empty(t, result, "empty builder should produce empty HCL")
}

// Test NewWithProvider()
func TestNewWithProvider(t *testing.T) {
	builder := hclbuilder.NewWithProvider(hclbuilder.KongMesh, "http://localhost:5681")
	result := builder.Build()

	goldenFile := filepath.Join("testdata", "new-with-provider.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test WithProvider()
func TestWithProvider(t *testing.T) {
	builder := hclbuilder.New()
	builder.WithProvider(hclbuilder.KongMesh, "http://localhost:5681")
	result := builder.Build()

	goldenFile := filepath.Join("testdata", "with-provider.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test SetAttribute() - simple
func TestSetAttribute_Simple(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetAttribute("resource.kong-mesh_mesh.default.name", "mesh-1")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "set-attribute-simple.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test SetAttribute() - multiple
func TestSetAttribute_Multiple(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetAttribute("resource.kong-mesh_mesh.default.name", "mesh-1")
	builder.SetAttribute("resource.kong-mesh_mesh.default.type", "Mesh")
	builder.SetAttribute("resource.kong-mesh_mesh_traffic_permission.allow_all.mesh", "kong-mesh_mesh.default.name")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "set-attribute-multiple.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test SetBlock() - simple
func TestSetBlock_Simple(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetBlock("resource.kong-mesh_mesh.default", map[string]any{
		"name":                           "mesh-1",
		"type":                           "Mesh",
		"skip_creating_initial_policies": []string{"*"},
	})

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "set-block-simple.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test SetBlock() - nested
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
	goldenFile := filepath.Join("testdata", "set-block-nested.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test RemoveAttribute()
func TestRemoveAttribute(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetAttribute("resource.kong-mesh_mesh.default.name", "mesh-1")
	builder.SetAttribute("resource.kong-mesh_mesh.default.type", "Mesh")
	builder.RemoveAttribute("resource.kong-mesh_mesh.default.type")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "remove-attribute.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test RemoveBlock()
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
	goldenFile := filepath.Join("testdata", "remove-block.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test FromFile()
func TestFromFile(t *testing.T) {
	inputFile := filepath.Join("testdata", "from-file.input.tf")
	builder, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)
	require.NotNil(t, builder)

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "from-file.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test FromFile() and modify loaded
func TestFromFile_ModifyLoaded(t *testing.T) {
	inputFile := filepath.Join("testdata", "from-file-modify.input.tf")
	builder, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	// Modify the loaded configuration
	builder.SetAttribute("resource.kong-mesh_mesh.default.skip_creating_initial_policies", []string{"*"})
	builder.SetAttribute("resource.kong-mesh_mesh_traffic_permission.existing.type", "MeshTrafficPermission")
	builder.SetBlock("resource.kong-mesh_mesh_traffic_permission.new_policy", map[string]any{
		"name": "new-policy",
	})

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "from-file-modify.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test WriteFile()
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

// Test FromString() - basic validation that strings can be loaded
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

// Test FromString() - invalid
func TestFromString_Invalid(t *testing.T) {
	invalidHCL := `resource "foo" {`
	_, err := hclbuilder.FromString(invalidHCL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing HCL")
}

// Test Add() - embed and mutate
func TestAdd_EmbedAndMutate(t *testing.T) {
	// Create a main file with a provider
	mainFile, err := hclbuilder.FromFile(filepath.Join("testdata", "add-embed-and-mutate-main.input.tf"))
	require.NoError(t, err)

	// Create a mesh resource
	mesh, err := hclbuilder.FromFile(filepath.Join("testdata", "add-embed-and-mutate-mesh.input.tf"))
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
	result := mainFile.Build()
	goldenFile := filepath.Join("testdata", "add-embed-and-mutate.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddAttribute() - nested
func TestAddAttribute_Nested(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-nested.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	mesh.AddAttribute("routing.default_forbid_mesh_external_service_access", true)

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-nested.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddAttribute() - remove simple
func TestAddAttribute_RemoveSimple(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-remove-simple.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	mesh.RemoveAttribute("skip_creating_initial_policies")

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-remove-simple.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddAttribute() - from HCL string simple list
func TestAddAttribute_FromHCLString_SimpleList(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-hcl-simple-list.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	mesh.AddAttribute("skip_creating_initial_policies", `["*"]`)

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-hcl-simple-list.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddAttribute() - from HCL string nested object
func TestAddAttribute_FromHCLString_NestedObject(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-hcl-nested-object.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	mesh.AddAttribute("constraints.dataplane_proxy.requirements", `[{ tags = { key = "a" } }]`)

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-hcl-nested-object.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddAttribute() - from HCL string boolean
func TestAddAttribute_FromHCLString_Boolean(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-hcl-boolean.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	mesh.AddAttribute("routing.default_forbid_mesh_external_service_access", "true")

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-hcl-boolean.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddAttribute() - from HCL string empty list
func TestAddAttribute_FromHCLString_EmptyList(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-hcl-empty-list.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	mesh.AddAttribute("constraints.dataplane_proxy.restrictions", "[]")

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-hcl-empty-list.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test ResourceAddress()
func TestResourceAddress(t *testing.T) {
	builder := hclbuilder.NewWithProvider(hclbuilder.KongMesh, "http://localhost:5681")

	address := builder.ResourceAddress("mesh", "default")
	require.Equal(t, "kong-mesh_mesh.default", address)
}

// Test ResourceAddress() - no provider
func TestResourceAddress_NoProvider(t *testing.T) {
	builder := hclbuilder.New()

	address := builder.ResourceAddress("mesh", "default")
	require.Equal(t, "mesh.default", address)
}

// Test RemoveMesh()
func TestRemoveMesh(t *testing.T) {
	builder := hclbuilder.NewWithProvider(hclbuilder.KongMesh, "http://localhost:5681")
	builder.SetBlock("resource.kong-mesh_mesh.default", map[string]any{
		"name": "mesh-1",
		"type": "Mesh",
	})
	builder.SetBlock("resource.kong-mesh_mesh.other", map[string]any{
		"name": "mesh-2",
		"type": "Mesh",
	})

	builder.RemoveMesh("default")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "remove-mesh.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddControlPlane()
func TestAddControlPlane(t *testing.T) {
	builder := hclbuilder.New()
	builder.ProviderProperty = hclbuilder.Konnect

	builder.AddControlPlane("cp1", "my-control-plane", "Test control plane")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "add-control-plane.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddPolicy()
func TestAddPolicy(t *testing.T) {
	builder := hclbuilder.NewWithProvider(hclbuilder.KongMesh, "http://localhost:5681")
	builder.ProviderProperty = hclbuilder.KongMesh

	builder.AddPolicy(
		"mesh_traffic_permission",
		"allow-all",
		"allow_all",
		"kong-mesh_mesh.default.name",
		map[string]any{
			"spec": map[string]any{
				"from": []any{
					map[string]any{
						"target_ref": map[string]any{
							"kind": "Mesh",
						},
					},
				},
			},
		},
	)

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "add-policy.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test ResourceName()
func TestResourceName(t *testing.T) {
	builder, err := hclbuilder.FromString(`
resource "kong-mesh_mesh" "my_mesh" {
  name = "mesh-1"
}
`)
	require.NoError(t, err)

	name := builder.ResourceName()
	require.Equal(t, "my_mesh", name)
}

// Test ResourceName() - empty
func TestResourceName_Empty(t *testing.T) {
	builder := hclbuilder.New()

	name := builder.ResourceName()
	require.Empty(t, name)
}

// Test Build() method
func TestBuild(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetBlock("resource.test.example", map[string]any{
		"name": "test",
	})

	result := builder.Build()
	require.Contains(t, result, "resource")
	require.Contains(t, result, "test")
	require.Contains(t, result, "example")
}

// Test Add() - adding same builder twice should only add once
func TestAdd_DuplicateBuilder(t *testing.T) {
	mainFile, err := hclbuilder.FromFile(filepath.Join("testdata", "add-duplicate-main.input.tf"))
	require.NoError(t, err)

	mesh, err := hclbuilder.FromFile(filepath.Join("testdata", "add-duplicate-mesh.input.tf"))
	require.NoError(t, err)

	// Add mesh twice
	mainFile.Add(mesh)
	mainFile.Add(mesh)

	// Should only be added once
	result := mainFile.Build()
	goldenFile := filepath.Join("testdata", "add-duplicate.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test Add() - merging attributes from other builder
func TestAdd_MergeAttributes(t *testing.T) {
	mainFile, err := hclbuilder.FromFile(filepath.Join("testdata", "add-merge-attributes-main.input.tf"))
	require.NoError(t, err)

	other, err := hclbuilder.FromFile(filepath.Join("testdata", "add-merge-attributes-other.input.tf"))
	require.NoError(t, err)

	mainFile.Add(other)

	result := mainFile.Build()
	goldenFile := filepath.Join("testdata", "add-merge-attributes.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddAttribute() - deep merge nested structures
func TestAddAttribute_DeepMerge(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-deep-merge.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	// Add a nested attribute that should merge with existing structure
	mesh.AddAttribute("constraints.dataplane_proxy.restrictions", []string{"restriction-1"})

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-deep-merge.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddAttribute() - parsing and converting existing expressions
func TestAddAttribute_ParseExistingExpression(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-parse-expression.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	// Add to existing nested structure by parsing existing expression
	mesh.AddAttribute("routing.locality_aware_load_balancing", false)

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-parse-expression.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test SetBlock() - with numeric values (int, int64, float64)
func TestSetBlock_NumericValues(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetBlock("resource.test.numeric", map[string]any{
		"int_value":    42,
		"int64_value":  int64(9223372036854775807),
		"float_value":  3.14159,
		"string_value": "test",
		"bool_value":   true,
	})

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "set-block-numeric-values.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test AddAttribute() - complex nested merge scenario
func TestAddAttribute_ComplexDeepMerge(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-complex-deep-merge.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	// First add to create nested structure
	mesh.AddAttribute("constraints.dataplane_proxy.requirements", `[{ tags = { key = "a" } }]`)
	// Then add another key to the same nested level
	mesh.AddAttribute("constraints.dataplane_proxy.restrictions", []string{"restriction-1"})

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-complex-deep-merge.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test ResourcePath()
func TestResourcePath(t *testing.T) {
	inputFile := filepath.Join("testdata", "resource-path.input.tf")
	builder, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	path := builder.ResourcePath()
	require.Equal(t, "konnect_mesh_control_plane.my_meshcontrolplane", path)
}

// Test ResourcePath() - empty builder
func TestResourcePath_Empty(t *testing.T) {
	builder := hclbuilder.New()

	path := builder.ResourcePath()
	require.Empty(t, path)
}

// Test ResourcePath() - non-resource block
func TestResourcePath_NonResource(t *testing.T) {
	inputFile := filepath.Join("testdata", "resource-path-non-resource.input.tf")
	builder, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	path := builder.ResourcePath()
	require.Empty(t, path)
}

// Test DependsOn() - single dependency
func TestDependsOn_Single(t *testing.T) {
	cpFile := filepath.Join("testdata", "depends-on-single-cp.input.tf")
	cp, err := hclbuilder.FromFile(cpFile)
	require.NoError(t, err)

	meshFile := filepath.Join("testdata", "depends-on-single-mesh.input.tf")
	mesh, err := hclbuilder.FromFile(meshFile)
	require.NoError(t, err)

	mesh.DependsOn(cp)

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "depends-on-single.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test DependsOn() - multiple dependencies
func TestDependsOn_Multiple(t *testing.T) {
	cpFile := filepath.Join("testdata", "depends-on-multiple-cp.input.tf")
	cp, err := hclbuilder.FromFile(cpFile)
	require.NoError(t, err)

	cp2File := filepath.Join("testdata", "depends-on-multiple-cp2.input.tf")
	cp2, err := hclbuilder.FromFile(cp2File)
	require.NoError(t, err)

	meshFile := filepath.Join("testdata", "depends-on-multiple-mesh.input.tf")
	mesh, err := hclbuilder.FromFile(meshFile)
	require.NoError(t, err)

	mesh.DependsOn(cp)
	mesh.DependsOn(cp2)

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "depends-on-multiple.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test DependsOn() - duplicate dependency should not add twice
func TestDependsOn_Duplicate(t *testing.T) {
	cpFile := filepath.Join("testdata", "depends-on-duplicate-cp.input.tf")
	cp, err := hclbuilder.FromFile(cpFile)
	require.NoError(t, err)

	meshFile := filepath.Join("testdata", "depends-on-duplicate-mesh.input.tf")
	mesh, err := hclbuilder.FromFile(meshFile)
	require.NoError(t, err)

	mesh.DependsOn(cp)
	mesh.DependsOn(cp)

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "depends-on-duplicate.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test DependsOn() - adding to existing depends_on
func TestDependsOn_ExistingDependencies(t *testing.T) {
	cpFile := filepath.Join("testdata", "depends-on-existing-cp.input.tf")
	cp, err := hclbuilder.FromFile(cpFile)
	require.NoError(t, err)

	meshFile := filepath.Join("testdata", "depends-on-existing-mesh.input.tf")
	mesh, err := hclbuilder.FromFile(meshFile)
	require.NoError(t, err)

	mesh.DependsOn(cp)

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "depends-on-existing.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestAddAttribute_ReferenceExpression_Simple(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-reference-simple.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	mesh.AddAttribute("cp_id", "konnect_mesh_control_plane.my_cp.id")

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-reference-simple.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestAddAttribute_ReferenceExpression_Nested(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-reference-nested.input.tf")
	policy, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	policy.AddAttribute("mesh_ref", "konnect_mesh.default.name")

	result := policy.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-reference-nested.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestAddAttribute_ReferenceExpression_WithIndex(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-reference-index.input.tf")
	resource, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	resource.AddAttribute("value", "data.aws_instances.example.ids[0]")

	result := resource.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-reference-index.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestAddAttribute_ReferenceExpression_MultiLevel(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-reference-multilevel.input.tf")
	resource, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	resource.AddAttribute("config", "module.networking.vpc.subnet.cidr_block")

	result := resource.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-reference-multilevel.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestAddAttribute_LiteralString_NotReference(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-reference-simple.input.tf")
	resource, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	// This is a literal string that happens to contain text, not a reference
	resource.AddAttribute("description", "This is just text")

	got := resource.Build()
	// Should be quoted since it's a literal string
	if !strings.Contains(got, `description = "This is just text"`) {
		t.Errorf("Expected literal string to be quoted, got:\n%s", got)
	}
}

func TestAddAttribute_ReferenceExpression_WithFunction(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-reference-function.input.tf")
	resource, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	resource.AddAttribute("timestamp", "timestamp()")

	result := resource.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-reference-function.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestAddAttribute_ReferenceExpression_Conditional(t *testing.T) {
	inputFile := filepath.Join("testdata", "add-attribute-reference-conditional.input.tf")
	resource, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	resource.AddAttribute("enabled", "var.environment == \"production\" ? true : false")

	result := resource.Build()
	goldenFile := filepath.Join("testdata", "add-attribute-reference-conditional.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}

// Test RemoveAttribute() - nested attribute removal
// This test demonstrates that RemoveAttribute should only remove the nested field,
// not the entire top-level attribute
func TestRemoveAttribute_Nested(t *testing.T) {
	inputFile := filepath.Join("testdata", "remove-attribute-nested.input.tf")
	mesh, err := hclbuilder.FromFile(inputFile)
	require.NoError(t, err)

	// Remove only the nested field, not the entire top-level attribute
	mesh.RemoveAttribute("routing.default_forbid_mesh_external_service_access")

	result := mesh.Build()
	goldenFile := filepath.Join("testdata", "remove-attribute-nested.golden.tf")
	assertGoldenFile(t, goldenFile, result)
}
