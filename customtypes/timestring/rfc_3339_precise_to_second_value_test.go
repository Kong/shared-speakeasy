package timetypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestRFC3339_Precise_to_Second_StringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentRFC3339time RFC3339PreciseToSecond
		givenRFC3339time   basetypes.StringValuable
		expectedMatch      bool
		expectedDiags      diag.Diagnostics
	}{
		"Semantically equal: Z suffix and positive zero local offset": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16Z"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16+00:00"),
			expectedMatch:      true,
		},
		"Semantically equal: Z suffix and negative zero local offset": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16Z"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16-00:00"),
			expectedMatch:      true,
		},
		"Semantically equal: negative zero and positive zero local offset": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16-00:00"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16+00:00"),
			expectedMatch:      true,
		},
		"Exacty equal": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16Z"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16Z"),
			expectedMatch:      true,
		},
		"Semantically equal: with different offsets for same time": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T21:43:16-02:00"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-16T01:43:16+02:00"),
			expectedMatch:      true,
		},
		"Semantically equal: with difference in milliseconds": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T21:43:16Z"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-15T21:43:16.101Z"),
			expectedMatch:      true,
		},
		"Not equal - different dates": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16Z"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-16T23:43:16Z"),
			expectedMatch:      false,
		},
		"Not equal - different times": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16Z"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-15T23:01:16Z"),
			expectedMatch:      false,
		},
		"Not equal - different offsets": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16Z"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16+03:00"),
			expectedMatch:      false,
		},
		"Not equal - UTC time and local time": {
			currentRFC3339time: NewRFC3339PreciseToSecondValueMust("2027-12-15T23:43:16Z"),
			givenRFC3339time:   NewRFC3339PreciseToSecondValueMust("2027-12-15T20:43:16-03:00"),
			expectedMatch:      true,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			match, _ := testCase.currentRFC3339time.StringSemanticEquals(context.Background(), testCase.givenRFC3339time)

			if testCase.expectedMatch != match {
				t.Errorf("Expected StringSemanticEquals to return: %t, but got: %t", testCase.expectedMatch, match)
			}
		})
	}
}
