package tfbuilder

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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

func ShouldBeAbleToStoreSecrets(providerFactory map[string]func() (tfprotov6.ProviderServer, error), builder *Builder, scert *PolicyBuilder, skey *PolicyBuilder, mesh *MeshBuilder) resource.TestCase {
	return resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		Steps: []resource.TestStep{
			{
				Config: builder.Build(),
			},
			{
				Config: builder.AddPolicy(scert.WithSpec("data = \"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvRENDQWVTZ0F3SUJBZ0lRUit3Y21qQ3pKNGtUcjdjMlRPeGVWekFOQmdrcWhraUc5dzBCQVFzRkFEQUEKTUI0WERUSTFNVEF5T1RFME1EWXdNbG9YRFRNMU1UQXlOekUwTURZd01sb3dBRENDQVNJd0RRWUpLb1pJaHZjTgpBUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTzNaOUlZTVRQMEhVT3FFbWptK0JOOWNrRlVBeVlGSjNsNjV3WE5xCjB3SFNzMGl4QVkrZHRoNmZQaGRqRmZ0eklYalhLbUYxLzBsRDdnSW5UN0J5RXB2SjBOWlRoVk9sYlZIZUR5ZXYKa1plOWYyTndTdTlJQnBQVjZqNGRpVFhuaFlkU3Y2MitxL2RyRzZXOFVaQmdweXh2TGRyS1FaYVlJUDFaVit4SgpxN2lYN2xzZDFJTHdjQ2wvVlY1MHBCcUJuT1FiejdEYnZ2OXVNUzYyZVFnSWRLRStsYmR6STcrSE53WU1KZzNzCi8rRkhBNGJNQUZseGZ6aURsS1MwMDNDazZHK3lTOUtoRWdscG1yOFhNeU85UVUvakhmOTNDOVVxOXBvbkhBZzgKVDAzcGlZbjRIWktlUkRlRENqM1lyL2JzZXZ4T1pDTmhNZmJLbnJnUElUMTVMS3NDQXdFQUFhTnlNSEF3RGdZRApWUjBQQVFIL0JBUURBZ0trTUJNR0ExVWRKUVFNTUFvR0NDc0dBUVVGQndNQk1BOEdBMVVkRXdFQi93UUZNQU1CCkFmOHdIUVlEVlIwT0JCWUVGRXcrYWw0UGpiZGRmeWo4UzRNNnBKSWxSN1Y2TUJrR0ExVWRFUUVCL3dRUE1BMkMKQzJWNFlXMXdiR1V1WTI5dE1BMEdDU3FHU0liM0RRRUJDd1VBQTRJQkFRRGpxcU5tcjlaSmFxaDErcUFkU0EzQgpxTHd0cUpOd2dvOEFHWFpiTFh4ajdkMTVsM3RuaDhzeHJ2Rkl6dzdiVVBlb1Q0MmhvN2xIaDdTUWFqVU9yR2dRCnZqYk5hSXRaZ3JRV2kzWEg4OEpFdC9aUWtKWEZlZ1hGYlBHL2VtMlBFTElncHdsYjNVVER0QWJXQjdEcFN4WDkKOE5TVzFFTGtEckI5VnFiWnM5ZWNMcDJUTmdRZmpsNG9ieVozRkxTVHlRN2I5c3ZHelk0a2VLelNUdzNld1B1MwpmS21ETFBTblg5U3BDRUREV2d1aHFaMzFsY3pXbzdUWmtYc0RsWkE5U3FndlZTTnIvM2s3UGVFSll1d00rL2syCnhxOHdiQUdmQkpBUmJab1FVd0NCS2duZm9rM0R3eHI5emZXc1JqUUsyeDBycFgrOFVuaWlmeXhmTzlDRHFqaDMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=\"")).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress(scert.ResourceType, scert.ResourceName), plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: builder.AddPolicy(skey.WithSpec("data = \"LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBN2RuMGhneE0vUWRRNm9TYU9iNEUzMXlRVlFESmdVbmVYcm5CYzJyVEFkS3pTTEVCCmo1MjJIcDgrRjJNViszTWhlTmNxWVhYL1NVUHVBaWRQc0hJU204blExbE9GVTZWdFVkNFBKNitSbDcxL1kzQksKNzBnR2s5WHFQaDJKTmVlRmgxSy9yYjZyOTJzYnBieFJrR0NuTEc4dDJzcEJscGdnL1ZsWDdFbXJ1SmZ1V3gzVQpndkJ3S1g5VlhuU2tHb0djNUJ2UHNOdSsvMjR4THJaNUNBaDBvVDZWdDNNanY0YzNCZ3dtRGV6LzRVY0Roc3dBCldYRi9PSU9VcExUVGNLVG9iN0pMMHFFU0NXbWF2eGN6STcxQlQrTWQvM2NMMVNyMm1pY2NDRHhQVGVtSmlmZ2QKa3A1RU40TUtQZGl2OXV4Ni9FNWtJMkV4OXNxZXVBOGhQWGtzcXdJREFRQUJBb0lCQUR0eG5HNGdCdUc2QVZ3TApOZXcyZEV0S2UvdnlqV25WaDFEUFJlek5odHpPeHVYazd3bndsWUtEcytYdWFxRUVQaHBRVkJRMWhFN1FQbHlsCmJJSWhrRXNGSGo5aWNsRGNhRHpzcllieWx3V0FZNlQ3Zkk3ZXhsNE9PVk82MS83ejFPaGtJdW1PWExZaU82K3AKS0ExWVNvK05YYjF2alFMUkZIV2M3Wjl0TGhDYzFKbWlaQUJpSVVPNG1ScW1vS3ZHRTd3N3NjQkFYRC8xZkdtVQp5QzlPMmNrd3BQN29HMGtmSFNUOGYzNFBKcTk1WHMreWNqdURFb1prT2Nzbnd4RVIzSzFYQnhjY1ZueXVHcU1KCkRKRTVreCtaVVRPQk5nemc4NEx3YmpwRG5BNTQwcW9xcHkwZVA3VTFKcHErOVByb25ZanpGNUxyUk9nZmJUbHEKMldhMGZ3RUNnWUVBL3dkdXpucDRzUmpkWmNGZGhMWlMyN1BYKzNmeVdsV0cvYXpWcHBXTzdTM1l1UXRLc2gwMApqVE9NY010cWloRWJHVGgzNXQwQStIamJCalpqYytPeHRITllkUFpDWml5N0hhaEV2Z1RVU1I0elRsVlVRMG14CitOWTY1Ym1KOU9PcXRiSnRDS3FMMnVlclF4UElta1UweGNkejVOaUJHQkJ6cWdEVUMvTm5iSUVDZ1lFQTdzSEgKcVdvaTJkZFNBaEdKU3dYOFp6Zit0dUxXT1NrNk5JR3dCNUpEM3VJUXlVaHdQbmxwM3BJNEhjQnh1WWIxK1BNego4MEdldGpXQ3dyemhTZWZ1bUtOb0VrQjB4eksrMlJsejYxTU1lWk1acUx2T2c3Mjg2WUxIV3NMRUVvUk5XRzN4Cmo0cVJaKzNvc1R6RFRNcVBSRVhGVDRzWnZ4bXRvaGxtZFdkY2N5c0NnWUVBakFyMDJnVit5U0V5VW5KQWZHUHkKVkJzSisxaitpSVIyd0U1c2RER2tickhDVkxyU3BjUkwyMDMzVE9rbTgvSTR3enl5K3Q5WmJSaFFqYlRJSUJkawp1Z2F0Q0cxQ1FRRkhMeDM3d2F5OU5mbVRpdXhvZlJxMjFFSXZ6WDU1TnpUZHhURFpsdXl3SitFWHRwbmlpblIrCmFpMEFneVl3blpwTEtZdVM1WTBmdWdFQ2dZQU52UEs3S2RORmk2RTVZejd1SlRNSDBXNERvZnZIb0Rxc0tNWXoKT1ZSVWI5ZWRiV0NnQjZaeTJ5RUZmVHhOKzVrTnNSak5KM3AxYTVEUm1jS3cyUHFlcDlCbU5IVkR2UVRFUXpXcgpWY1VDL2RiZElhbFpaVUtJZ1REdFpRV1pOeW1vSy9OWldoVFIwUnV4anhpQnc2b0l1S2NJMDYwd2xNNnI1Q0JFCkl5VnJyd0tCZ0I3emJsVUZwZXVPNlFXSUNCMEpQZWRDbjBjL3B0blhHdFVuTVNVN2FCTWhINGVYWWJXUmwyZTgKc0g2bHhUVC82MmVVYk1hM0l5VWcyQmdsaEx3UWpjdzNIVngvUVJZMHBTQWRpSlVGWXhSczFrVU1KTDBLMk15WApsbVdiQzZ5b05HREZ0ZGY1aFdULzRTNDBXNVZmWkRjZS9MZytlei9tZ0hVeUFCNU9vOE1RCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==\"")).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress(skey.ResourceType, skey.ResourceName), plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: builder.AddMesh(mesh.WithSpec(`skip_creating_initial_policies = [ "*" ]
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
`)).Build(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(builder.ResourceAddress("mesh", mesh.ResourceName), plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: builder.AddMesh(mesh.WithSpec(`skip_creating_initial_policies = [ "*" ]`)).Build(),
			},
			CheckReapplyPlanEmpty(builder),
		},
	}
}
