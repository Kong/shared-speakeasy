package tfbuilder

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
)

type debugPlan struct{}

func (e debugPlan) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
    reqPlan, err := json.Marshal(req.Plan)
    if err != nil {
        tflog.Debug(ctx, fmt.Sprintf("error marshaling plan request: %s", err))
    }
    tflog.Info(ctx, fmt.Sprintf("req.Plan - %s\n", string(reqPlan)))
}

func DebugPlan() plancheck.PlanCheck {
    return debugPlan{}
}
