package tfbuilder

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func CheckReapplyPlanEmpty(builder *Builder) resource.TestStep {
	return resource.TestStep{
		// Re-apply the same config and ensure no changes occur
		Config: builder.Build(),
		ConfigPlanChecks: resource.ConfigPlanChecks{
			PreApply: []plancheck.PlanCheck{
				plancheck.ExpectEmptyPlan(),
			},
		},
	}
}
