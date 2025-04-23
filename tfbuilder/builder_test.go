package tfbuilder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Kong/shared-speakeasy/tfbuilder"
)

func TestBuilder_KongMeshWithPolicy(t *testing.T) {
	builder := tfbuilder.NewBuilder(tfbuilder.KongMesh, "http", "localhost", 5681)

	// Add mesh
	builder.AddMesh(
		tfbuilder.NewMeshBuilder("default", "mesh-1").
			WithSpec(`skip_creating_initial_policies = [ "*" ]`),
	)

	// Add policy
	builder.AddPolicy(
		tfbuilder.NewPolicyBuilder("mesh_traffic_permission", "allow_all", "allow-all", "MeshTrafficPermission").
			WithMeshRef("kong-mesh_mesh.default.name").
			WithDependsOn("kong-mesh_mesh.default").
			WithLabels(map[string]string{
				"kuma.io/mesh": "kong-mesh_mesh.default.name",
			}).
			WithSpecHCL(tfbuilder.AllowAllTrafficPermissionSpec),
	)

	actual := builder.Build()

	goldenFile := filepath.Join("testdata", "expected_kong_mesh_with_policy.tf")

	if updateGoldenFiles(t, goldenFile, actual) {
		return
	}

	expected, err := os.ReadFile(goldenFile)
	require.NoError(t, err, "reading golden file")

	require.Equal(t, string(expected), actual)
}

func updateGoldenFiles(t *testing.T, goldenFile string, actual string) bool {
	if os.Getenv("UPDATE_GOLDEN_FILES") == "1" {
		err := os.WriteFile(goldenFile, []byte(actual), 0o600)
		require.NoError(t, err, "updating golden file")
		t.Logf("updated golden file: %s", goldenFile)
		return true
	}
	return false
}
