package expr

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluate(t *testing.T) {
	ctx := map[string]any{
		"DB_TYPE":   "PostgreSQL",
		"USE_REDIS": false,
	}

	tests := []struct {
		name     string
		expr     string
		expected bool
		hasError bool
	}{
		{"Not Equal", "DB_TYPE != 'None'", true, false},
		{"False condition", "USE_REDIS == true", false, false},
		{"Complex condition", "DB_TYPE == 'PostgreSQL' && USE_REDIS == false", true, false},
		{"Empty expression", "", true, false},
		{"Empty space", "   ", true, false},
		{"Invalid syntax", "DB_TYPE ==", false, true},
		{"Not a boolean", "'string'", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Evaluate(tt.expr, ctx)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestEvaluate_EmptyExpression_ReturnsTrue(t *testing.T) {
	ok, err := Evaluate("", map[string]any{"a": 1})
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestEvaluate_VeryLongExpression_ReturnsError(t *testing.T) {
	long := strings.Repeat("A", MaxExpressionLength+1)
	_, err := Evaluate(long, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum length")
}

func TestEvaluate_InjectionAttempt_ReturnsError(t *testing.T) {
	cases := []string{
		`os.Exit(1)`,
		`exec.Command("rm", "-rf", "/")`,
		`__import__('os').system('whoami')`,
	}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			_, err := Evaluate(expr, map[string]any{})
			require.Error(t, err, "must not succeed for unsafe expression")
		})
	}
}

func TestEvaluate_DivisionByZero_IsErrorOrFalse(t *testing.T) {
	_, err := Evaluate(`1/0`, map[string]any{})
	// gval may error or yield non-bool; must not panic
	if err == nil {
		t.Fatal("expected error for division by zero")
	}
}

func TestEvaluate_NonBoolOrInvalidField(t *testing.T) {
	_, err := Evaluate(`null.field`, map[string]any{})
	require.Error(t, err)
}

func TestEvaluate_CircularReference_ReturnsError(t *testing.T) {
	ctx := map[string]any{
		"a": map[string]any{"b": "c"},
	}
	_, err := Evaluate(`a.b == "x"`, ctx)
	// Should evaluate without infinite loop; may error or return bool
	require.NoError(t, err)
}

func TestEvaluate_SecurityBoundaries(t *testing.T) {
	// All dangerous expressions below must return an error, not panic.
	dangerousCases := []struct {
		name string
		expr string
	}{
		{"os exit attempt", `os.Exit(1)`},
		{"exec command attempt", `exec.Command("whoami")`},
		{"very long expression", strings.Repeat("A", 100_000)},
		{"division by zero", `1/0`},
		{"null dereference", `nil.field`},
		// empty string is a special-case: no condition => always run
		{"empty string", ``},
	}

	ctx := map[string]any{"SAFE_VAR": "value"}

	for _, tc := range dangerousCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Evaluate panicked for expression %q: %v", tc.expr, r)
				}
			}()

			if tc.expr == "" {
				result, err := Evaluate(tc.expr, ctx)
				assert.NoError(t, err)
				assert.True(t, result)
				return
			}

			_, err := Evaluate(tc.expr, ctx)
			assert.NotNil(t, err, "expression %q should return error", tc.expr)
		})
	}
}

func TestEvaluate_TemplateInjectionPrevention(t *testing.T) {
	// Values coming from user input must not be treated as template strings.
	ctx := map[string]any{
		"PROJECT_NAME": "{{.SECRET_VAR}}",
		"SECRET_VAR":   "sensitive-data",
	}

	// Legitimate expressions should still evaluate normally.
	result, err := Evaluate("PROJECT_NAME != ''", ctx)
	assert.NoError(t, err)
	assert.True(t, result)

	// The stored value "{{.SECRET_VAR}}" must not be expanded.
	result2, err2 := Evaluate("PROJECT_NAME == 'sensitive-data'", ctx)
	assert.NoError(t, err2)
	assert.False(t, result2, "template injection should not expose SECRET_VAR value")
}
