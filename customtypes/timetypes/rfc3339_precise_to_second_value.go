package timetypes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable    = (*RFC3339PreciseToSecond)(nil)
	_ xattr.ValidateableAttribute = (*RFC3339PreciseToSecond)(nil)
)

// RFC3339PreciseToSecond represents a valid RFC3339-formatted string. It supports second-level precision and ignores milliseconds.
// Semantic equality for this type normalizes timestamps to UTC, truncates the milliseconds, and compares the results.
type RFC3339PreciseToSecond struct {
	basetypes.StringValue
}

// Type returns an RFC3339PreciseToSecondType.
func (v RFC3339PreciseToSecond) Type(_ context.Context) attr.Type {
	return RFC3339PreciseToSecondType{}
}

// Equal returns true if the given value is equivalent.
func (v RFC3339PreciseToSecond) Equal(o attr.Value) bool {
	other, ok := o.(RFC3339PreciseToSecond)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals returns true if the given RFC3339 timestamp is semantically equal the current RFC3339 timestamp.
// This comparison ignores milliseconds, and normalizes both timestamps to UTC before comparison.
//
// Examples:
//   - `2023-07-25T22:43:16+02:00` is semantically equal to `2023-07-25T20:43:16Z`
//   - `2023-07-25T20:43:16.05Z` is semantically equal to `2023-07-25T20:43:16Z`
func (v RFC3339PreciseToSecond) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(RFC3339PreciseToSecond)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	newRFC3339time, _ := time.Parse(time.RFC3339, newValue.ValueString())
	currentRFC3339time, _ := time.Parse(time.RFC3339, v.ValueString())

	return normalize(currentRFC3339time) == normalize(newRFC3339time), diags
}

// Helper function to normalize a time.Time to UTC and truncate to second precision, and return as RFC3339 string.
func normalize(t time.Time) string {
	return t.UTC().Truncate(time.Second).Format(time.RFC3339)
}

// This type requires the value to be a String value in valid RFC 3339 format
func (v RFC3339PreciseToSecond) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsUnknown() || v.IsNull() {
		return
	}

	if _, err := time.Parse(time.RFC3339, v.ValueString()); err != nil {
		resp.Diagnostics.Append(
			diag.WithPath(req.Path, diag.NewErrorDiagnostic("Invalid attribute value", "The attribute value must be a string in RFC3339 format.")),
		)

		return
	}
}

// NewRFC3339PreciseToSecondNull creates an RFC3339PreciseToSecond with a null value. Determine whether the value is null via IsNull method.
func NewRFC3339PreciseToSecondNull() RFC3339PreciseToSecond {
	return RFC3339PreciseToSecond{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewRFC3339PreciseToSecondUnknown creates an RFC3339PreciseToSecond with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewRFC3339PreciseToSecondUnknown() RFC3339PreciseToSecond {
	return RFC3339PreciseToSecond{
		StringValue: basetypes.NewStringUnknown(),
	}
}

// NewRFC3339PreciseToSecondValue creates an RFC3339PreciseToSecond with a known value or raises an error
// diagnostic if the string is not RFC3339 format.
func NewRFC3339PreciseToSecondValue(value string) (RFC3339PreciseToSecond, diag.Diagnostics) {
	_, err := time.Parse(time.RFC3339, value)
	if err != nil {
		// Returning an unknown value will guarantee that, as a last resort,
		// Terraform will return an error if attempting to store into state.
		return NewRFC3339PreciseToSecondUnknown(), diag.Diagnostics{}
	}

	return RFC3339PreciseToSecond{
		StringValue: basetypes.NewStringValue(value),
	}, nil
}

// NewRFC3339PreciseToSecondValueMust creates an RFC3339PreciseToSecond with a known value or raises a panic
// if the string is not RFC3339 format.
// Used in unit tests.
func NewRFC3339PreciseToSecondValueMust(value string) RFC3339PreciseToSecond {
	_, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(fmt.Sprintf("Invalid RFC3339 String Value (%s): %s", value, err))
	}

	return RFC3339PreciseToSecond{
		StringValue: basetypes.NewStringValue(value),
	}
}

// NewRFC3339PreciseToSecondPointerValue creates an RFC3339PreciseToSecond with a null value if nil, a known
// value, or raises an error diagnostic if the string is not RFC3339 format.
func NewRFC3339PreciseToSecondPointerValue(value *string) (RFC3339PreciseToSecond, diag.Diagnostics) {
	if value == nil {
		return NewRFC3339PreciseToSecondNull(), nil
	}

	return NewRFC3339PreciseToSecondValue(*value)
}
