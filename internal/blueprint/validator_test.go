package blueprint

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidation_Required(t *testing.T) {
	rule := ValidationRule{Field: "NAME", Rule: "required", Message: "Harap diisi."}
	rules := []ValidationRule{rule}

	// 1. Valid: String tidak kosong
	errs := ValidateInput("NAME", "MyApp", rules)
	assert.Empty(t, errs)

	// 2. Invalid: String kosong
	errs = ValidateInput("NAME", "", rules)
	assert.Len(t, errs, 1)
	assert.Equal(t, "Harap diisi.", errs[0].Message)

	// 3. Invalid: Spasi saja
	errs = ValidateInput("NAME", "   ", rules)
	assert.Len(t, errs, 1)

	// 4. Invalid: nil
	errs = ValidateInput("NAME", nil, rules)
	assert.Len(t, errs, 1)

	// 5. Invalid: slice kosong
	errs = ValidateInput("NAME", []string{}, rules)
	assert.Len(t, errs, 1)
}

func TestValidation_Regex(t *testing.T) {
	rule := ValidationRule{Field: "EMAIL", Rule: "regex", Pattern: `^[a-z]+@[a-z]+\.[a-z]+$`}
	rules := []ValidationRule{rule}

	// 1. Valid
	errs := ValidateInput("EMAIL", "test@test.com", rules)
	assert.Empty(t, errs)

	// 2. Invalid
	errs = ValidateInput("EMAIL", "bukan-email", rules)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "tidak memenuhi format")

	// 3. String kosong (aturan regex tidak mengcover input kosong; pakai required untuk itu!)
	errs = ValidateInput("EMAIL", "", rules)
	assert.Empty(t, errs)

	// 4. Invalid pattern konfigurasi template
	invalidRule := ValidationRule{Field: "BAD", Rule: "regex", Pattern: `[`}
	errs = ValidateInput("BAD", "anything", []ValidationRule{invalidRule})
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "regex tidak valid")
}

func TestValidation_MinLength(t *testing.T) {
	rule := ValidationRule{Field: "PASS", Rule: "min_length", Pattern: "8"}
	rules := []ValidationRule{rule}

	// 1. Valid
	errs := ValidateInput("PASS", "12345678", rules)
	assert.Empty(t, errs)

	// 2. Invalid
	errs = ValidateInput("PASS", "1234567", rules)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "Minimal 8 karakter")
}

func TestValidation_MaxLength(t *testing.T) {
	rule := ValidationRule{Field: "USERNAME", Rule: "max_length", Pattern: "10"}
	rules := []ValidationRule{rule}

	// 1. Valid
	errs := ValidateInput("USERNAME", "ninechars", rules)
	assert.Empty(t, errs)

	// 2. Invalid
	errs = ValidateInput("USERNAME", "elevenchars", rules)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "Maksimal 10 karakter")
}

func TestValidation_MultipleRules(t *testing.T) {
	rules := []ValidationRule{
		{Field: "TOKEN", Rule: "required"},
		{Field: "TOKEN", Rule: "min_length", Pattern: "5"},
		{Field: "TOKEN", Rule: "max_length", Pattern: "10"},
	}

	errs := ValidateInput("TOKEN", "abc", rules)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "Minimal 5")
	assert.Equal(t, "min_length", errs[0].Rule)

	errs2 := ValidateInput("TOKEN", "", rules)
	// Output harus 'required' dan 'min_length' keduanya error
	assert.Len(t, errs2, 2)
}
