package hclbuilder

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// MeshConfig holds mesh configuration
type MeshConfig struct {
	MeshName     string
	ResourceName string
	Provider     ProviderType
}

// PolicyConfig holds policy configuration
type PolicyConfig struct {
	PolicyType   string
	PolicyName   string
	ResourceName string
	MeshRef      string
	Provider     ProviderType
}

// CheckReapplyPlanEmpty verifies that reapplying produces an empty plan
func CheckReapplyPlanEmpty(builder interface{ Build() string }) resource.TestStep {
	return resource.TestStep{
		Config:             builder.Build(),
		PlanOnly:           true,
		ExpectNonEmptyPlan: false,
	}
}

// CreateMeshAndModifyFields creates a mesh and modifies fields on it
func CreateMeshAndModifyFields(
	providerFactory map[string]func() (tfprotov6.ProviderServer, error),
	meshConfig MeshConfig,
) resource.TestCase {
	builder := NewTestBuilder(meshConfig.Provider)
	meshResourcePath := builder.ResourceAddress("mesh", meshConfig.ResourceName)

	// Initial spec
	initialSpec := map[string]any{
		"skip_creating_initial_policies": []string{"*"},
	}

	// Spec with additional fields
	extendedSpec := map[string]any{
		"skip_creating_initial_policies": []string{"*"},
		"constraints": map[string]any{
			"dataplane_proxy": map[string]any{
				"requirements": []map[string]any{
					{"tags": map[string]any{"key": "a"}},
				},
				"restrictions": []any{},
			},
		},
		"routing": map[string]any{
			"default_forbid_mesh_external_service_access": true,
		},
	}

	// Spec with routing field removed
	withoutRoutingSpec := map[string]any{
		"skip_creating_initial_policies": []string{"*"},
		"constraints": map[string]any{
			"dataplane_proxy": map[string]any{
				"requirements": []map[string]any{
					{"tags": map[string]any{"key": "a"}},
				},
				"restrictions": []any{},
			},
		},
	}

	// Spec with requirements updated
	updatedRequirementsSpec := map[string]any{
		"skip_creating_initial_policies": []string{"*"},
		"constraints": map[string]any{
			"dataplane_proxy": map[string]any{
				"requirements": []any{},
				"restrictions": []any{},
			},
		},
	}

	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.AddMesh(meshConfig.MeshName, meshConfig.ResourceName, initialSpec).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionCreate),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.AddMesh(meshConfig.MeshName, meshConfig.ResourceName, extendedSpec).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(meshResourcePath, tfjsonpath.New("routing").AtMapKey("default_forbid_mesh_external_service_access"), knownvalue.Bool(true)),
					},
				},
			},
			{
				Config: builder.AddMesh(meshConfig.MeshName, meshConfig.ResourceName, withoutRoutingSpec).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(meshResourcePath, tfjsonpath.New("routing").AtMapKey("default_forbid_mesh_external_service_access"), knownvalue.Null()),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.AddMesh(meshConfig.MeshName, meshConfig.ResourceName, updatedRequirementsSpec).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(meshResourcePath, tfjsonpath.New("constraints").AtMapKey("dataplane_proxy").AtMapKey("requirements"), knownvalue.ListExact([]knownvalue.Check{})),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.RemoveMesh(meshConfig.ResourceName).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(meshResourcePath, plancheck.ResourceActionDestroy),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
		},
	}
}

// CreatePolicyAndModifyFields creates a policy and modifies fields on it
func CreatePolicyAndModifyFields(
	providerFactory map[string]func() (tfprotov6.ProviderServer, error),
	policyConfig PolicyConfig,
) resource.TestCase {
	builder := NewTestBuilder(policyConfig.Provider)
	policyResourcePath := builder.ResourceAddress("mesh_traffic_permission", policyConfig.ResourceName)

	// Initial spec
	initialSpec := map[string]any{
		"spec": map[string]any{
			"from": []map[string]any{
				{
					"target_ref": map[string]any{
						"kind": "Mesh",
					},
					"default": map[string]any{},
				},
			},
		},
	}

	// Spec with proxy_types added
	withProxyTypesSpec := map[string]any{
		"spec": map[string]any{
			"from": []map[string]any{
				{
					"target_ref": map[string]any{
						"kind":        "Mesh",
						"proxy_types": []string{"Sidecar"},
					},
					"default": map[string]any{},
				},
			},
		},
	}

	// Spec with proxy_types as empty array
	emptyProxyTypesSpec := map[string]any{
		"spec": map[string]any{
			"from": []map[string]any{
				{
					"target_ref": map[string]any{
						"kind":        "Mesh",
						"proxy_types": []string{},
					},
					"default": map[string]any{},
				},
			},
		},
	}

	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.AddPolicy(policyConfig.PolicyType, policyConfig.PolicyName, policyConfig.ResourceName, policyConfig.MeshRef, initialSpec).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(policyResourcePath, plancheck.ResourceActionCreate),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.AddPolicy(policyConfig.PolicyType, policyConfig.PolicyName, policyConfig.ResourceName, policyConfig.MeshRef, withProxyTypesSpec).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(policyResourcePath, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(policyResourcePath, tfjsonpath.New("spec").AtMapKey("from").AtSliceIndex(0).AtMapKey("target_ref").AtMapKey("proxy_types"), knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("Sidecar")})),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.AddPolicy(policyConfig.PolicyType, policyConfig.PolicyName, policyConfig.ResourceName, policyConfig.MeshRef, emptyProxyTypesSpec).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(policyResourcePath, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(policyResourcePath, tfjsonpath.New("spec").AtMapKey("from").AtSliceIndex(0).AtMapKey("target_ref").AtMapKey("proxy_types"), knownvalue.ListExact([]knownvalue.Check{})),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
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
	builder := NewTestBuilder(policyConfig.Provider)
	policyResourcePath := builder.ResourceAddress("mesh_traffic_permission", policyConfig.ResourceName)

	spec := map[string]any{
		"spec": map[string]any{
			"from": []map[string]any{
				{
					"target_ref": map[string]any{
						"kind": "Mesh",
					},
					"default": map[string]any{},
				},
			},
		},
	}

	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.Build(),
			},
			{
				PreConfig:   preConfigFn,
				Config:      builder.AddPolicy(policyConfig.PolicyType, policyConfig.PolicyName, policyConfig.ResourceName, policyConfig.MeshRef, spec).Build(),
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
