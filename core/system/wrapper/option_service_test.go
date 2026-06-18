package wrapper

import (
	"context"
	"testing"
)

type fakeOptionObjectService map[string]map[string]any

func (f fakeOptionObjectService) GetObjectByName(_ context.Context, name string) (map[string]any, error) {
	if value, ok := f[name]; ok {
		return value, nil
	}
	return nil, assertErr("missing option")
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}

func TestPasswordPolicyReadsCurrentSystemSecurityShape(t *testing.T) {
	wrapper := &OptionServiceWrapper{
		options: fakeOptionObjectService{
			optionSystemSecurity: {
				"passwordMinLength":  12,
				"passwordComplexity": false,
				"requireSymbols":     true,
			},
		},
	}

	policy := wrapper.PasswordPolicy(context.Background())
	if policy.MinLength != 12 {
		t.Fatalf("expected min length 12, got %d", policy.MinLength)
	}
	if policy.RequireUppercase || policy.RequireLowercase || policy.RequireNumbers {
		t.Fatalf("expected complexity flags disabled, got %#v", policy)
	}
	if !policy.RequireSymbols {
		t.Fatalf("expected symbols to be required, got %#v", policy)
	}
}

func TestPasswordPolicySupportsExplicitSnakeCaseOverrides(t *testing.T) {
	wrapper := &OptionServiceWrapper{
		options: fakeOptionObjectService{
			optionSystemSecurity: {
				"min_length":         10,
				"passwordComplexity": true,
				"require_numbers":    false,
				"require_symbols":    true,
			},
		},
	}

	policy := wrapper.PasswordPolicy(context.Background())
	if policy.MinLength != 10 {
		t.Fatalf("expected min length 10, got %d", policy.MinLength)
	}
	if !policy.RequireUppercase || !policy.RequireLowercase {
		t.Fatalf("expected complexity to require upper and lower case, got %#v", policy)
	}
	if policy.RequireNumbers {
		t.Fatalf("expected explicit number override, got %#v", policy)
	}
	if !policy.RequireSymbols {
		t.Fatalf("expected explicit symbol override, got %#v", policy)
	}
}
