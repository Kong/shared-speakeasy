package timetypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.StringTypable = (*RFC3339PreciseToSecondType)(nil)

// RFC3339PreciseToSecondType is an attribute type that represents a valid RFC 3339 string, but only to second precision.
// This type is useful for timestamps that don't require millisecond precision.
type RFC3339PreciseToSecondType struct {
	basetypes.StringType
}

// String returns a human-readable string of the type name.
func (t RFC3339PreciseToSecondType) String() string {
	return "timetypes.RFC3339PreciseToSecondType"
}

// ValueType returns the Value type.
func (t RFC3339PreciseToSecondType) ValueType(ctx context.Context) attr.Value {
	return RFC3339PreciseToSecond{}
}

// Equal returns true if the given type is equivalent.
func (t RFC3339PreciseToSecondType) Equal(o attr.Type) bool {
	other, ok := o.(RFC3339PreciseToSecondType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t RFC3339PreciseToSecondType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return RFC3339PreciseToSecond{
		StringValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t RFC3339PreciseToSecondType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
