package engine

import (
	"fmt"

	"github.com/username/symphony/internal/blueprint"
)

// ValidateInputs menerapkan blueprint.Validations terhadap nilai di ctx.Values
// menggunakan metode validasi yang tersentralisasi di package blueprint.
func ValidateInputs(bp *blueprint.Blueprint, ctx *EngineContext) error {
	var allErrs []error

	for _, rule := range bp.Validations {
		field := rule.Field
		val := ctx.Get(field)

		errs := blueprint.ValidateInput(field, val, []blueprint.ValidationRule{rule})
		for _, e := range errs {
			allErrs = append(allErrs, e)
		}
	}

	if len(allErrs) > 0 {
		return fmt.Errorf("validasi gagal dengan %d error. Pertama: %v", len(allErrs), allErrs[0])
	}

	return nil
}
func valueAsString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}
