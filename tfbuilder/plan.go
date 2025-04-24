package tfbuilder

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "log"
)

type debugPlan struct{}

func (e debugPlan) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
    reqPlan, err := json.Marshal(req.Plan)
    if err != nil {
        log.Fatalf(fmt.Sprintf("error marshaling plan request: %s", err))
    }
    reqPlanStr := string(reqPlan)
    log.Printf("req.Plan - %s\n", reqPlanStr)
}

func DebugPlan() plancheck.PlanCheck {
    return debugPlan{}
}
