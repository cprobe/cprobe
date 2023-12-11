package exporter

import (
	"math"
	"testing"
)

func TestSanitizeValue(t *testing.T) {
	tests := []struct {
		Input          string
		ExpectedOutput float64
		ShouldSucceed  bool
	}{
		{"1234", 1234.0, true},
		{"1234.5", 1234.5, true},
		{"true", 1.0, true},
		{"TRUE", 1.0, true},
		{"False", 0.0, true},
		{"FALSE", 0.0, true},
		{"abcd", 0, false},
		{"{}", 0, false},
		{"[]", 0, false},
		{"", 0, false},
		{"''", 0, false},
	}

	for i, test := range tests {
		actualOutput, err := SanitizeValue(test.Input)
		if err != nil && test.ShouldSucceed {
			t.Fatalf("Value snitization test %d failed with an unexpected error.\nINPUT:\n%q\nERR:\n%s", i, test.Input, err)
		}
		if test.ShouldSucceed && actualOutput != test.ExpectedOutput {
			t.Fatalf("Value sanitization test %d fails unexpectedly.\nGOT:\n%f\nEXPECTED:\n%f", i, actualOutput, test.ExpectedOutput)
		}
	}
}

func TestSanitizeValueNaN(t *testing.T) {
	actualOutput, err := SanitizeValue("<nil>")
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(actualOutput) {
		t.Fatalf("Value sanitization test for %f fails unexpectedly.", math.NaN())
	}
}
