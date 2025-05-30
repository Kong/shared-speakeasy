package tfbuilder

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"regexp"
)

func CreateMeshAndModifyFieldsOnIt(
	providerFactory map[string]func() (tfprotov6.ProviderServer, error),
	builder ModifyMeshBuilder,
	mesh *MeshBuilder,
) resource.TestCase {
	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.AddMesh(mesh).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress("mesh", mesh.ResourceName), plancheck.ResourceActionCreate),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.AddMesh(mesh.AddToSpec("at the end", `
  constraints = {
    dataplane_proxy = {
      requirements = [ { tags = { key = "a" } } ]
      restrictions = []
    }
  }
  routing = {
    default_forbid_mesh_external_service_access = true
  }
`)).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress("mesh", mesh.ResourceName), plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(builder.ResourceAddress("mesh", mesh.ResourceName), tfjsonpath.New("routing").AtMapKey("default_forbid_mesh_external_service_access"), knownvalue.Bool(true)),
					},
				},
			},
			{
				Config: builder.AddMesh(mesh.RemoveFromSpec(`default_forbid_mesh_external_service_access = true`)).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress("mesh", mesh.ResourceName), plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(builder.ResourceAddress("mesh", mesh.ResourceName), tfjsonpath.New("routing").AtMapKey("default_forbid_mesh_external_service_access"), knownvalue.Null()),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.AddMesh(mesh.UpdateSpec(`requirements = [ { tags = { key = "a" } } ]`, `requirements = []`)).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress("mesh", mesh.ResourceName), plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(builder.ResourceAddress("mesh", mesh.ResourceName), tfjsonpath.New("constraints").AtMapKey("dataplane_proxy").AtMapKey("requirements"), knownvalue.ListExact([]knownvalue.Check{})),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.RemoveMesh(mesh.MeshName).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress("mesh", mesh.ResourceName), plancheck.ResourceActionDestroy),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
		},
	}
}

func CreatePolicyAndModifyFieldsOnIt(
	providerFactory map[string]func() (tfprotov6.ProviderServer, error),
	builder ModifyPolicyBuilder,
	mtp *PolicyBuilder,
) resource.TestCase {
	mtp.WithSpec(AllowAllTrafficPermissionSpec)

	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.AddPolicy(mtp).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress(mtp.ResourceType, mtp.ResourceName), plancheck.ResourceActionCreate),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.AddPolicy(mtp.AddToSpec(`kind = "Mesh"`, `proxy_types = ["Sidecar"]`)).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress(mtp.ResourceType, mtp.ResourceName), plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(builder.ResourceAddress(mtp.ResourceType, mtp.ResourceName), tfjsonpath.New("spec").AtMapKey("from").AtSliceIndex(0).AtMapKey("target_ref").AtMapKey("proxy_types"), knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("Sidecar")})),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
			{
				Config: builder.AddPolicy(mtp.UpdateSpec(`proxy_types = ["Sidecar"]`, `proxy_types = []`)).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress(mtp.ResourceType, mtp.ResourceName), plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(builder.ResourceAddress(mtp.ResourceType, mtp.ResourceName), tfjsonpath.New("spec").AtMapKey("from").AtSliceIndex(0).AtMapKey("target_ref").AtMapKey("proxy_types"), knownvalue.ListExact([]knownvalue.Check{})),
					},
				},
			},
			CheckReapplyPlanEmpty(builder),
		},
	}
}

func NotImportedResourceShouldErrorOutWithMeaningfulMessage(
	providerFactory map[string]func() (tfprotov6.ProviderServer, error),
	builder ModifyPolicyBuilder,
	mtp *PolicyBuilder,
	preConfigFn func(),
) resource.TestCase {
	expectedErr := regexp.MustCompile(`MeshTrafficPermission already exists`)

	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.Build(),
			},
			{
				PreConfig:   preConfigFn,
				Config:      builder.AddPolicy(mtp.WithSpec(AllowAllTrafficPermissionSpec)).Build(),
				ExpectError: expectedErr,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress(mtp.ResourceType, mtp.ResourceName), plancheck.ResourceActionCreate),
					},
				},
			},
		},
	}
}
