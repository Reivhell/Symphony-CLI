package expr

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/gval"
)

// MaxExpressionLength limits blueprint "if" expressions to mitigate ReDoS and abuse.
const MaxExpressionLength = 4096

// Evaluate evaluates a boolean expression against context values.
// Panics from the underlying evaluator are converted to errors.
func Evaluate(expression string, context map[string]any) (result bool, err error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return true, nil
	}

	if len(expression) > MaxExpressionLength {
		return false, fmt.Errorf("expression exceeds maximum length (%d bytes)", MaxExpressionLength)
	}

	expression = strings.ReplaceAll(expression, "'", "\"")

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("expression evaluation failed: %v", r)
			result = false
		}
	}()

	val, evalErr := gval.Evaluate(expression, context)
	if evalErr != nil {
		return false, fmt.Errorf("failed to evaluate expression: %w", evalErr)
	}

	b, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("expression must evaluate to a boolean, got %T", val)
	}

	return b, nil
}
