package engine

import (
	"testing"

	"github.com/username/symphony/internal/blueprint"
	"github.com/stretchr/testify/assert"
)

func TestValidateInputs(t *testing.T) {
	bp := &blueprint.Blueprint{
		Validations: []blueprint.ValidationRule{
			{Field: "PROJECT_NAME", Rule: "required", Message: "nama wajib"},
			{Field: "SLUG", Rule: "regex", Pattern: `^[a-z0-9-]+$`, Message: "slug invalid"},
		},
	}
	ctx := &EngineContext{Values: map[string]any{"PROJECT_NAME": "", "SLUG": "ok-slug"}}
	err := ValidateInputs(bp, ctx)
	assert.Error(t, err)

	ctx.Values["PROJECT_NAME"] = "app"
	err = ValidateInputs(bp, ctx)
	assert.NoError(t, err)

	ctx.Values["SLUG"] = "Bad Slug"
	err = ValidateInputs(bp, ctx)
	assert.Error(t, err)
}
