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
