package kumalabels

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EmptyKumaLabelsMapDefault struct{}

func (EmptyKumaLabelsMapDefault) Description(context.Context) string {
	return "defaults to empty map"
}

func (EmptyKumaLabelsMapDefault) MarkdownDescription(ctx context.Context) string {
	return "defaults to empty map"
}

func (EmptyKumaLabelsMapDefault) DefaultMap(ctx context.Context, _ defaults.MapRequest, resp *defaults.MapResponse) {
	resp.PlanValue = types.MapValueMust(types.StringType, map[string]attr.Value{})
}
