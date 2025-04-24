package tfbuilder

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

type debugPlan struct{}

func (e debugPlan) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	reqPlan, err := json.Marshal(req.Plan)
	if err != nil {
		log.Fatalf("error marshaling plan request: %s", err)
	}
	reqPlanStr := string(reqPlan)
	log.Printf("req.Plan - %s\n", reqPlanStr)
}

func DebugPlan() plancheck.PlanCheck {
	return debugPlan{}
}
