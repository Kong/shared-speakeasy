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
						plancheck.ExpectKnownValue(meshResourcePath, tfjsonpath.New("routing"), knownvalue.Null()),
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

	mesh, _ := FromString(fmt.Sprintf(`
resource "%s_mesh" "%s" {
  type  = "Mesh"
  name  = "%s"
  skip_creating_initial_policies = ["*"]
}
`, builder.ProviderPrefix(), meshResourceName, meshName))
	builder.Add(mesh)

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

	mesh, _ := FromString(fmt.Sprintf(`
resource "%s_mesh" "%s" {
  type  = "Mesh"
  name  = "%s"
  skip_creating_initial_policies = ["*"]
}
`, builder.ProviderPrefix(), meshResourceName, meshName))
	builder.Add(mesh)

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

// ShouldBeAbleToStoreSecrets tests storing and using secrets with a mesh
func ShouldBeAbleToStoreSecrets(
	providerFactory map[string]func() (tfprotov6.ProviderServer, error),
	secretConfig PolicyConfig,
) resource.TestCase {
	builder := NewWithProvider(string(secretConfig.Provider), secretConfig.ServerURL)

	// Create mesh
	meshResourceName := "test_mesh"
	meshName := secretConfig.MeshRef

	mesh, _ := FromString(fmt.Sprintf(`
resource "kong-mesh_mesh" "%s" {
  type = "Mesh"
  name = "%s"
  skip_creating_initial_policies = ["*"]
}
`, meshResourceName, meshName))
	builder.Add(mesh)

	// Create secret resources
	scertResourceName := "scert"
	scertName := "scert"
	skeyResourceName := "skey"
	skeyName := "skey"

	meshResourceAddress := builder.ResourceAddress("mesh", meshResourceName)
	scertResourceAddress := builder.ResourceAddress("mesh_secret", scertResourceName)
	skeyResourceAddress := builder.ResourceAddress("mesh_secret", skeyResourceName)

	// Step 1 config: just mesh
	step1Config := builder.Build()

	// Step 2: Add cert secret
	scertSpec := mustParseSpec(`
		labels = {}
		depends_on = []
		data = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvRENDQWVTZ0F3SUJBZ0lRUit3Y21qQ3pKNGtUcjdjMlRPeGVWekFOQmdrcWhraUc5dzBCQVFzRkFEQUEKTUI0WERUSTFNVEF5T1RFME1EWXdNbG9YRFRNMU1UQXlOekUwTURZd01sb3dBRENDQVNJd0RRWUpLb1pJaHZjTgpBUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTzNaOUlZTVRQMEhVT3FFbWptK0JOOWNrRlVBeVlGSjNsNjV3WE5xCjB3SFNzMGl4QVkrZHRoNmZQaGRqRmZ0eklYalhLbUYxLzBsRDdnSW5UN0J5RXB2SjBOWlRoVk9sYlZIZUR5ZXYKa1plOWYyTndTdTlJQnBQVjZqNGRpVFhuaFlkU3Y2MitxL2RyRzZXOFVaQmdweXh2TGRyS1FaYVlJUDFaVit4SgpxN2lYN2xzZDFJTHdjQ2wvVlY1MHBCcUJuT1FiejdEYnZ2OXVNUzYyZVFnSWRLRStsYmR6STcrSE53WU1KZzNzCi8rRkhBNGJNQUZseGZ6aURsS1MwMDNDazZHK3lTOUtoRWdscG1yOFhNeU85UVUvakhmOTNDOVVxOXBvbkhBZzgKVDAzcGlZbjRIWktlUkRlRENqM1lyL2JzZXZ4T1pDTmhNZmJLbnJnUElUMTVMS3NDQXdFQUFhTnlNSEF3RGdZRApWUjBQQVFIL0JBUURBZ0trTUJNR0ExVWRKUVFNTUFvR0NDc0dBUVVGQndNQk1BOEdBMVVkRXdFQi93UUZNQU1CCkFmOHdIUVlEVlIwT0JCWUVGRXcrYWw0UGpiZGRmeWo4UzRNNnBKSWxSN1Y2TUJrR0ExVWRFUUVCL3dRUE1BMkMKQzJWNFlXMXdiR1V1WTI5dE1BMEdDU3FHU0liM0RRRUJDd1VBQTRJQkFRRGpxcU5tcjlaSmFxaDErcUFkU0EzQgpxTHd0cUpOd2dvOEFHWFpiTFh4ajdkMTVsM3RuaDhzeHJ2Rkl6dzdiVVBlb1Q0MmhvN2xIaDdTUWFqVU9yR2dRCnZqYk5hSXRaZ3JRV2kzWEg4OEpFdC9aUWtKWEZlZ1hGYlBHL2VtMlBFTElncHdsYjNVVER0QWJXQjdEcFN4WDkKOE5TVzFFTGtEckI5VnFiWnM5ZWNMcDJUTmdRZmpsNG9ieVozRkxTVHlRN2I5c3ZHelk0a2VLelNUdzNld1B1MwpmS21ETFBTblg5U3BDRUREV2d1aHFaMzFsY3pXbzdUWmtYc0RsWkE5U3FndlZTTnIvM2s3UGVFSll1d00rL2syCnhxOHdiQUdmQkpBUmJab1FVd0NCS2duZm9rM0R3eHI5emZXc1JqUUsyeDBycFgrOFVuaWlmeXhmTzlDRHFqaDMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	`)
	scertSpec["depends_on"] = []any{meshResourceAddress}
	builder.AddPolicy("mesh_secret", scertName, scertResourceName, meshName, scertSpec)
	step2Config := builder.Build()

	// Step 3: Add key secret
	skeySpec := mustParseSpec(`
		labels = {}
		depends_on = []
		data = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBN2RuMGhneE0vUWRRNm9TYU9iNEUzMXlRVlFESmdVbmVYcm5CYzJyVEFkS3pTTEVCCmo1MjJIcDgrRjJNViszTWhlTmNxWVhYL1NVUHVBaWRQc0hJU204blExbE9GVTZWdFVkNFBKNitSbDcxL1kzQksKNzBnR2s5WHFQaDJKTmVlRmgxSy9yYjZyOTJzYnBieFJrR0NuTEc4dDJzcEJscGdnL1ZsWDdFbXJ1SmZ1V3gzVQpndkJ3S1g5VlhuU2tHb0djNUJ2UHNOdSsvMjR4THJaNUNBaDBvVDZWdDNNanY0YzNCZ3dtRGV6LzRVY0Roc3dBCldYRi9PSU9VcExUVGNLVG9iN0pMMHFFU0NXbWF2eGN6STcxQlQrTWQvM2NMMVNyMm1pY2NDRHhQVGVtSmlmZ2QKa3A1RU40TUtQZGl2OXV4Ni9FNWtJMkV4OXNxZXVBOGhQWGtzcXdJREFRQUJBb0lCQUR0eG5HNGdCdUc2QVZ3TApOZXcyZEV0S2UvdnlqV25WaDFEUFJlek5odHpPeHVYazd3bndsWUtEcytYdWFxRUVQaHBRVkJRMWhFN1FQbHlsCmJJSWhrRXNGSGo5aWNsRGNhRHpzcllieWx3V0FZNlQ3Zkk3ZXhsNE9PVk82MS83ejFPaGtJdW1PWExZaU82K3AKS0ExWVNvK05YYjF2alFMUkZIV2M3Wjl0TGhDYzFKbWlaQUJpSVVPNG1ScW1vS3ZHRTd3N3NjQkFYRC8xZkdtVQp5QzlPMmNrd3BQN29HMGtmSFNUOGYzNFBKcTk1WHMreWNqdURFb1prT2Nzbnd4RVIzSzFYQnhjY1ZueXVHcU1KCkRKRTVreCtaVVRPQk5nemc4NEx3YmpwRG5BNTQwcW9xcHkwZVA3VTFKcHErOVByb25ZanpGNUxyUk9nZmJUbHEKMldhMGZ3RUNnWUVBL3dkdXpucDRzUmpkWmNGZGhMWlMyN1BYKzNmeVdsV0cvYXpWcHBXTzdTM1l1UXRLc2gwMApqVE9NY010cWloRWJHVGgzNXQwQStIamJCalpqYytPeHRITllkUFpDWml5N0hhaEV2Z1RVU1I0elRsVlVRMG14CitOWTY1Ym1KOU9PcXRiSnRDS3FMMnVlclF4UElta1UweGNkejVOaUJHQkJ6cWdEVUMvTm5iSUVDZ1lFQTdzSEgKcVdvaTJkZFNBaEdKU3dYOFp6Zit0dUxXT1NrNk5JR3dCNUpEM3VJUXlVaHdQbmxwM3BJNEhjQnh1WWIxK1BNego4MEdldGpXQ3dyemhTZWZ1bUtOb0VrQjB4eksrMlJsejYxTU1lWk1acUx2T2c3Mjg2WUxIV3NMRUVvUk5XRzN4Cmo0cVJaKzNvc1R6RFRNcVBSRVhGVDRzWnZ4bXRvaGxtZFdkY2N5c0NnWUVBakFyMDJnVit5U0V5VW5KQWZHUHkKVkJzSisxaitpSVIyd0U1c2RER2tickhDVkxyU3BjUkwyMDMzVE9rbTgvSTR3enl5K3Q5WmJSaFFqYlRJSUJkawp1Z2F0Q0cxQ1FRRkhMeDM3d2F5OU5mbVRpdXhvZlJxMjFFSXZ6WDU1TnpUZHhURFpsdXl3SitFWHRwbmlpblIrCmFpMEFneVl3blpwTEtZdVM1WTBmdWdFQ2dZQU52UEs3S2RORmk2RTVZejd1SlRNSDBXNERvZnZIb0Rxc0tNWXoKT1ZSVWI5ZWRiV0NnQjZaeTJ5RUZmVHhOKzVrTnNSak5KM3AxYTVEUm1jS3cyUHFlcDlCbU5IVkR2UVRFUXpXcgpWY1VDL2RiZElhbFpaVUtJZ1REdFpRV1pOeW1vSy9OWldoVFIwUnV4anhpQnc2b0l1S2NJMDYwd2xNNnI1Q0JFCkl5VnJyd0tCZ0I3emJsVUZwZXVPNlFXSUNCMEpQZWRDbjBjL3B0blhHdFVuTVNVN2FCTWhINGVYWWJXUmwyZTgKc0g2bHhUVC82MmVVYk1hM0l5VWcyQmdsaEx3UWpjdzNIVngvUVJZMHBTQWRpSlVGWXhSczFrVU1KTDBLMk15WApsbVdiQzZ5b05HREZ0ZGY1aFdULzRTNDBXNVZmWkRjZS9MZytlei9tZ0hVeUFCNU9vOE1RCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
	`)
	skeySpec["depends_on"] = []any{meshResourceAddress}
	builder.AddPolicy("mesh_secret", skeyName, skeyResourceName, meshName, skeySpec)
	step3Config := builder.Build()

	// Step 4: Add mesh with mtls configuration using the secrets
	meshWithMTLS, _ := FromString(fmt.Sprintf(`
resource "kong-mesh_mesh" "%s" {
  type = "Mesh"
  name = "%s"
  skip_creating_initial_policies = ["*"]
  mtls = {
    backends = [
      {
        conf = {
          provided_certificate_authority_config = {
            cert = {
              data_source_secret = {
                secret = "scert"
              }
            }
            key = {
              data_source_secret = {
                secret = "skey"
              }
            }
          }
        }
        name = "provided-1"
        type = "provided"
      }
    ]
    enabled_backend = "provided-1"
  }
}
`, meshResourceName, meshName))

	// Remove the old mesh and add the updated one
	builder.RemoveMesh(meshResourceName)
	builder.Add(meshWithMTLS)
	step4Config := builder.Build()

	// Step 5: Remove mtls from mesh
	mesh2, _ := FromString(fmt.Sprintf(`
resource "kong-mesh_mesh" "%s" {
  type = "Mesh"
  name = "%s"
  skip_creating_initial_policies = ["*"]
}
`, meshResourceName, meshName))
	builder.RemoveMesh(meshResourceName)
	builder.Add(mesh2)
	step5Config := builder.Build()

	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: step1Config,
			},
			{
				Config: step2Config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(scertResourceAddress, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: step3Config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(skeyResourceAddress, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: step4Config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourceAddress, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: step5Config,
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
