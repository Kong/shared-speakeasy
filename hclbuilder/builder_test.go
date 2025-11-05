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
	builder.SetAttribute("variable.name.default", "test-value")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "simple_attribute.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestSetAttribute_Multiple(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetAttribute("variable.name.default", "my-name")
	builder.SetAttribute("variable.name.description", "A test variable")
	builder.SetAttribute("variable.count.default", 5)

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "multiple_attributes.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestSetBlock_Simple(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetBlock("resource.aws_instance.web", map[string]any{
		"ami":           "ami-123456",
		"instance_type": "t2.micro",
	})

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "simple_block.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestSetBlock_Nested(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetBlock("resource.aws_security_group.example", map[string]any{
		"name": "example",
		"ingress": map[string]any{
			"from_port":   80,
			"to_port":     80,
			"protocol":    "tcp",
			"cidr_blocks": []string{"0.0.0.0/0"},
		},
	})

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "nested_block.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestRemoveAttribute(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetAttribute("variable.name.default", "test")
	builder.SetAttribute("variable.name.sensitive", true)
	builder.RemoveAttribute("variable.name.sensitive")

	result := builder.Build()
	goldenFile := filepath.Join("testdata", "removed_attribute.tf")
	assertGoldenFile(t, goldenFile, result)
}

func TestRemoveBlock(t *testing.T) {
	builder := hclbuilder.New()
	builder.SetBlock("resource.aws_instance.web", map[string]any{
		"ami": "ami-123",
	})
	builder.SetBlock("resource.aws_instance.old", map[string]any{
		"ami": "ami-old",
	})
	builder.RemoveBlock("resource.aws_instance.old")

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
	builder.SetAttribute("variable.new_var.default", "added")
	builder.RemoveBlock("resource.aws_instance.old")

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
