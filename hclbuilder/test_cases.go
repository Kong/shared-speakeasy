package hclbuilder

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
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

// MeshConfig holds mesh configuration
type MeshConfig struct {
	MeshName     string
	ResourceName string
	Provider     ProviderType
	ServerURL    string // Optional server URL, defaults to http://localhost:5681
}

// PolicyConfig holds policy configuration
type PolicyConfig struct {
	PolicyType   string
	PolicyName   string
	ResourceName string
	MeshRef      string
	Provider     ProviderType
	ServerURL    string // Optional server URL, defaults to http://localhost:5681
}

// CreateMeshAndModifyFields creates a mesh and modifies fields on it
func CreateMeshAndModifyFields(
	providerFactory map[string]func() (tfprotov6.ProviderServer, error),
	meshConfig MeshConfig,
) resource.TestCase {
	builder := NewWithProvider(string(meshConfig.Provider), meshConfig.ServerURL)
	meshResourcePath := builder.ResourceAddress("mesh", meshConfig.ResourceName)

	mesh, _ := FromString(fmt.Sprintf(`
resource "kong-mesh_mesh" "%s" {
  type  = "Mesh"
  name  = "%s"
  skip_creating_initial_policies = ["*"]
}
`, meshConfig.ResourceName, meshConfig.MeshName))

	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.Add(mesh).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: builder.Add(mesh.
					AddAttribute("constraints.dataplane_proxy.requirements", `[{ tags = { key = "a" } }]`).
					AddAttribute("constraints.dataplane_proxy.restrictions", `[]`).
					AddAttribute("routing.default_forbid_mesh_external_service_access", "true")).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(meshResourcePath, tfjsonpath.New("routing").AtMapKey("default_forbid_mesh_external_service_access"), knownvalue.Bool(true)),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: builder.Add(mesh.
					RemoveAttribute("routing")).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(meshResourcePath, tfjsonpath.New("routing").AtMapKey("default_forbid_mesh_external_service_access"), knownvalue.Null()),
					},
				},
			},
			{
				Config: builder.Add(mesh.
					AddAttribute("constraints.dataplane_proxy.requirements", "[]")).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(meshResourcePath, tfjsonpath.New("constraints").AtMapKey("dataplane_proxy").AtMapKey("requirements"), knownvalue.ListExact([]knownvalue.Check{})),
					},
				},
			},
			{
				Config: builder.RemoveMesh(meshConfig.ResourceName).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionDestroy),
					},
				},
			},
		},
	}
}

// CreatePolicyAndModifyFields creates a policy and modifies fields on it
func CreatePolicyAndModifyFields(
	providerFactory map[string]func() (tfprotov6.ProviderServer, error),
	policyConfig PolicyConfig,
) resource.TestCase {
	builder := NewWithProvider(string(policyConfig.Provider), policyConfig.ServerURL)

	// Create mesh first for the policy tests
	meshResourceName := "test_mesh"
	meshName := "policy-test-mesh"
	builder.AddMesh(meshName, meshResourceName, mustParseSpec(`
		skip_creating_initial_policies = ["*"]
	`))

	policyResourcePath := builder.ResourceAddress("mesh_traffic_permission", policyConfig.ResourceName)
	meshResourceAddress := builder.ResourceAddress("mesh", meshResourceName)

	policySpec := mustParseSpec(`
		labels = {}
		spec = {
			from = [{
				target_ref = {
					kind = "Mesh"
					proxy_types = ["Sidecar"]
				}
				default = {
					action = "Allow"
				}
			}]
		}
	`)
	policySpec["depends_on"] = []any{meshResourceAddress}

	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.AddPolicy(policyConfig.PolicyType, policyConfig.PolicyName, policyConfig.ResourceName, meshName, policySpec).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(policyResourcePath, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: builder.Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(policyResourcePath, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(policyResourcePath, tfjsonpath.New("spec").AtMapKey("from").AtSliceIndex(0).AtMapKey("target_ref").AtMapKey("proxy_types"), knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("Sidecar")})),
					},
				},
			},
			{
				Config: builder.Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(policyResourcePath, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(policyResourcePath, tfjsonpath.New("spec").AtMapKey("from").AtSliceIndex(0).AtMapKey("target_ref").AtMapKey("proxy_types"), knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("Sidecar")})),
					},
				},
			},
		},
	}
}

// NotImportedResourceShouldError tests error handling for non-imported resources
func NotImportedResourceShouldError(
	providerFactory map[string]func() (tfprotov6.ProviderServer, error),
	policyConfig PolicyConfig,
	preConfigFn func(),
) resource.TestCase {
	expectedErr := regexp.MustCompile(`MeshTrafficPermission already exists`)
	builder := NewWithProvider(string(policyConfig.Provider), policyConfig.ServerURL)

	// Create mesh first
	meshResourceName := "test_mesh"
	meshName := "policy-test-mesh-2"
	builder.AddMesh(meshName, meshResourceName, mustParseSpec(`
		skip_creating_initial_policies = ["*"]
	`))

	policyResourcePath := builder.ResourceAddress("mesh_traffic_permission", policyConfig.ResourceName)
	meshResourceAddress := builder.ResourceAddress("mesh", meshResourceName)

	policySpec := mustParseSpec(`
		labels = {}
		spec = {
			from = [{
				target_ref = {
					kind = "Mesh"
					proxy_types = ["Sidecar"]
				}
				default = {
					action = "Allow"
				}
			}]
		}
	`)
	policySpec["depends_on"] = []any{meshResourceAddress}

	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.Build(),
			},
			{
				PreConfig:   preConfigFn,
				Config:      builder.AddPolicy(policyConfig.PolicyType, policyConfig.PolicyName, policyConfig.ResourceName, meshName, policySpec).Build(),
				ExpectError: expectedErr,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(policyResourcePath, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	}
}

// mustParseSpec parses HCL attributes from a string and returns them as a map.
// It uses FromString to validate syntax, then hclparse to evaluate expressions.
func mustParseSpec(hclAttrs string) map[string]any {
	// Wrap in a dummy block to make it valid HCL
	wrapped := fmt.Sprintf("dummy {\n%s\n}", hclAttrs)

	// Validate syntax using FromString (uses hclwrite for syntax check)
	_, err := FromString(wrapped)
	if err != nil {
		panic(fmt.Sprintf("invalid HCL syntax: %s", err))
	}

	// Parse and evaluate using hclparse (needed for expression evaluation)
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(wrapped), "<inline>")
	if diags.HasErrors() {
		panic(fmt.Sprintf("failed to parse: %s", diags.Error()))
	}

	// Extract the dummy block
	bodyContent, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "dummy"}},
	})
	if diags.HasErrors() || len(bodyContent.Blocks) == 0 {
		panic("failed to extract dummy block")
	}

	// Get attributes from the dummy block
	attrs, diags := bodyContent.Blocks[0].Body.JustAttributes()
	if diags.HasErrors() {
		panic(fmt.Sprintf("failed to get attributes: %s", diags.Error()))
	}

	result := make(map[string]any)
	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			panic(fmt.Sprintf("failed to evaluate %s: %s", name, diags.Error()))
		}
		result[name] = ctyToGo(val)
	}

	return result
}

// ctyToGo converts a cty.Value to a Go value
func ctyToGo(val cty.Value) any {
	if val.IsNull() {
		return nil
	}

	ty := val.Type()

	switch {
	case ty == cty.String:
		return val.AsString()
	case ty == cty.Number:
		var f float64
		_ = gocty.FromCtyValue(val, &f)
		// Check if it's actually an integer
		if f == float64(int64(f)) {
			return int64(f)
		}
		return f
	case ty == cty.Bool:
		return val.True()
	case ty.IsListType() || ty.IsTupleType():
		var result []any
		it := val.ElementIterator()
		for it.Next() {
			_, elemVal := it.Element()
			result = append(result, ctyToGo(elemVal))
		}
		return result
	case ty.IsMapType() || ty.IsObjectType():
		result := make(map[string]any)
		it := val.ElementIterator()
		for it.Next() {
			keyVal, elemVal := it.Element()
			key := keyVal.AsString()
			result[key] = ctyToGo(elemVal)
		}
		return result
	default:
		panic(fmt.Sprintf("unsupported cty type: %s", ty.FriendlyName()))
	}
}
