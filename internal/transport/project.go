package transport

import "fmt"

func ProjectIndicatorVectors(indicators []IndicatorSpec, cases []ScenarioResult) []IndicatorVector {
	result := make([]IndicatorVector, 0, len(indicators))
	for _, indicator := range indicators {
		decision := Closed
		for _, item := range cases {
			if item.Ordinal == indicator.Ordinal {
				decision = item.Decision
				break
			}
		}
		result = append(result, IndicatorVector{
			Ordinal: indicator.Ordinal,
			ID: indicator.ID,
			Family: indicator.Family,
			Role: indicator.Role,
			Activity: indicator.Activity,
			Decision: decision,
			Numerator: boolInt(decision == Closed),
			Denominator: 1,
		})
	}
	return result
}

func ProjectSemanticIR(ir SemanticIR) (SemanticIR, error) {
	if err := ValidateSemanticIR(ir); err != nil {
		return SemanticIR{}, err
	}
	ir.ExpectedAssets = sortAssets(ir.ExpectedAssets)
	return ir, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func ValidateVectors(cases []CaseVector, indicators []IndicatorVector, denominator int) error {
	if len(cases) != denominator || len(indicators) != 12 {
		return fmt.Errorf("vector lengths do not match declared denominators")
	}
	for index, item := range cases {
		if item.Ordinal != index+1 || item.Denominator != 1 || (item.Numerator != 0 && item.Numerator != 1) {
			return fmt.Errorf("case vector %d is not an exact 1-cell vector", index+1)
		}
	}
	for index, item := range indicators {
		if item.Ordinal != index+1 || item.Denominator != 1 || (item.Numerator != 0 && item.Numerator != 1) {
			return fmt.Errorf("indicator vector %d is not an exact 1-cell vector", index+1)
		}
	}
	return nil
}
