package hclbuilder

import (
	"fmt"
	"strings"
)

// ProviderType represents a Terraform provider type
type ProviderType string

const (
	// KongMesh is the kong-mesh provider
	KongMesh ProviderType = "kong-mesh"
	// Konnect is the konnect provider
	Konnect ProviderType = "konnect"
	// KonnectBeta is the konnect-beta provider
	KonnectBeta ProviderType = "konnect-beta"
)

// TestBuilder wraps Builder with provider-specific helpers for test cases
type TestBuilder struct {
	builder      *Builder
	providerType ProviderType
}

// NewTestBuilder creates a new test builder with a provider configuration
func NewTestBuilder(provider ProviderType) *TestBuilder {
	b := New()

	// Add provider block
	providerName := strings.ToLower(string(provider))
	providerName = strings.ReplaceAll(providerName, "_", "-")

	b.SetBlock(fmt.Sprintf("provider.%s", providerName), map[string]any{
		"server_url": "http://localhost:5681",
	})

	return &TestBuilder{
		builder:      b,
		providerType: provider,
	}
}

// Build returns the HCL configuration string
func (tb *TestBuilder) Build() string {
	return tb.builder.Build()
}

// ResourceAddress returns the Terraform resource address
func (tb *TestBuilder) ResourceAddress(resourceType, resourceName string) string {
	providerPrefix := tb.getProviderPrefix()
	return fmt.Sprintf("%s_%s.%s", providerPrefix, resourceType, resourceName)
}

// AddMeshFromHCL adds a mesh resource from an HCL block
func (tb *TestBuilder) AddMeshFromHCL(meshName, meshResourceName, hclBlock string) *TestBuilder {
	providerPrefix := tb.getProviderPrefix()

	attrs := map[string]any{
		"provider": strings.ToLower(string(tb.providerType)),
		"type":     "Mesh",
		"name":     meshName,
	}

	// Parse the HCL block to get attributes if not empty
	if hclBlock != "" {
		// Use mustParseSpec from test_cases.go to parse the attributes
		// Wrap in a function call since mustParseSpec is in test_cases.go
		parsed := parseHCLAttributes(hclBlock)
		for k, v := range parsed {
			attrs[k] = v
		}
	}

	meshPath := fmt.Sprintf("resource.%s_mesh.%s", providerPrefix, meshResourceName)
	tb.builder.SetBlock(meshPath, attrs)

	return tb
}

// AddMesh adds a mesh resource with a spec map
func (tb *TestBuilder) AddMesh(meshName, meshResourceName string, spec map[string]any) *TestBuilder {
	providerPrefix := tb.getProviderPrefix()

	attrs := map[string]any{
		"provider": strings.ToLower(string(tb.providerType)),
		"type":     "Mesh",
		"name":     meshName,
	}

	// Merge spec into attrs
	for k, v := range spec {
		attrs[k] = v
	}

	tb.builder.SetBlock(fmt.Sprintf("resource.%s_mesh.%s", providerPrefix, meshResourceName), attrs)
	return tb
}

// RemoveMesh removes a mesh resource
func (tb *TestBuilder) RemoveMesh(meshResourceName string) *TestBuilder {
	providerPrefix := tb.getProviderPrefix()
	tb.builder.RemoveBlock(fmt.Sprintf("resource.%s_mesh.%s", providerPrefix, meshResourceName))
	return tb
}

// AddPolicy adds a policy resource with a spec map
func (tb *TestBuilder) AddPolicy(policyType, policyName, policyResourceName, meshRef string, spec map[string]any) *TestBuilder {
	providerPrefix := tb.getProviderPrefix()

	// Convert policy type from snake_case to PascalCase for the type attribute
	pascalCaseType := tb.resourceTypeToPolicyType(policyType)

	attrs := map[string]any{
		"provider": strings.ToLower(string(tb.providerType)),
		"type":     pascalCaseType,
		"name":     policyName,
		"mesh":     meshRef,
	}

	// Merge spec into attrs
	for k, v := range spec {
		attrs[k] = v
	}

	tb.builder.SetBlock(fmt.Sprintf("resource.%s_%s.%s", providerPrefix, policyType, policyResourceName), attrs)
	return tb
}

// RemovePolicy removes a policy resource
func (tb *TestBuilder) RemovePolicy(policyType, policyResourceName string) *TestBuilder {
	providerPrefix := tb.getProviderPrefix()
	resourceType := tb.policyTypeToResourceType(policyType)
	tb.builder.RemoveBlock(fmt.Sprintf("resource.%s_%s.%s", providerPrefix, resourceType, policyResourceName))
	return tb
}

// SetAttribute sets an attribute on a resource
func (tb *TestBuilder) SetAttribute(resourcePath, attribute string, value any) *TestBuilder {
	tb.builder.SetAttribute(fmt.Sprintf("%s.%s", resourcePath, attribute), value)
	return tb
}

// RemoveAttribute removes an attribute from a resource
func (tb *TestBuilder) RemoveAttribute(resourcePath, attribute string) *TestBuilder {
	tb.builder.RemoveAttribute(fmt.Sprintf("%s.%s", resourcePath, attribute))
	return tb
}

func (tb *TestBuilder) getProviderPrefix() string {
	prefix := strings.ToLower(string(tb.providerType))
	return strings.ReplaceAll(prefix, "_", "-")
}

func (tb *TestBuilder) policyTypeToResourceType(policyType string) string {
	// Convert "MeshTrafficPermission" to "mesh_traffic_permission"
	result := ""
	for i, r := range policyType {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += strings.ToLower(string(r))
	}
	return result
}

func (tb *TestBuilder) resourceTypeToPolicyType(resourceType string) string {
	// Convert "mesh_traffic_permission" to "MeshTrafficPermission"
	parts := strings.Split(resourceType, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	return result
}

// parseHCLAttributes parses HCL attributes from a string
func parseHCLAttributes(hclAttrs string) map[string]any {
	// Call mustParseSpec from test_cases.go (same package)
	return mustParseSpec(hclAttrs)
}
