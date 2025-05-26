package kumalabels

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.MapTypable = KumaLabelsMapType{}

// taken from https://github.com/kumahq/kuma/blob/99d5d12d1a67c76ec3d4f992486900b970f85a10/pkg/core/resources/model/labels/labels.go#L8
var AllComputedLabels = map[string]struct{}{
	"kuma.io/mesh":          {},
	"kuma.io/origin":        {},
	"kuma.io/zone":          {},
	"kuma.io/env":           {},
	"k8s.kuma.io/namespace": {},
	"kuma.io/policy-role":   {},
	"kuma.io/proxy-type":    {},
}

type KumaLabelsMapType struct {
	basetypes.MapType
}

func (t KumaLabelsMapType) String() string {
	return "KumaLabelsMapType"
}

func (t KumaLabelsMapType) Equal(o attr.Type) bool {
	other, ok := o.(KumaLabelsMapType)
	if !ok {
		return false
	}
	return t.MapType.Equal(other.MapType)
}

func (t KumaLabelsMapType) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	filteredElements := make(map[string]attr.Value)
	for key, val := range in.Elements() {
		if _, ok := AllComputedLabels[key]; !ok {
			filteredElements[key] = val
		}
	}

	filteredMapValue, diags := types.MapValue(types.StringType, filteredElements)
	if diags.HasError() {
		return nil, diags
	}

	return KumaLabelsMapValue{MapValue: filteredMapValue}, nil
}

func (t KumaLabelsMapType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.MapType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	mapValue, ok := attrValue.(basetypes.MapValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	val, diags := t.ValueFromMap(ctx, mapValue)
	if diags.HasError() {
		return nil, fmt.Errorf("error filtering KumaLabelsMapType: %v", diags)
	}

	return val, nil
}

func (t KumaLabelsMapType) ValueType(ctx context.Context) attr.Value {
	return KumaLabelsMapValue{}
}

type KumaLabelsMapValue struct {
	basetypes.MapValue
}

func (t KumaLabelsMapValue) Type(ctx context.Context) attr.Type {
	return KumaLabelsMapType{
		MapType: types.MapType{
			ElemType: types.StringType,
		},
	}
}

func (v KumaLabelsMapValue) Equal(o attr.Value) bool {
	other, ok := o.(KumaLabelsMapValue)
	if !ok {
		return false
	}
	return v.MapValue.Equal(other.MapValue)
}
