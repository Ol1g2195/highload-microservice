package validation

import (
	"testing"
)

type sampleStruct struct {
	Password string `validate:"strong_password"`
	Safe     string `validate:"safe_string"`
	UUID     string `validate:"uuid"`
	Email    string `validate:"email_domain"`
	SQL      string `validate:"no_sql_injection"`
	XSS      string `validate:"no_xss"`
}

func mustValidator(t *testing.T) *CustomValidator {
	t.Helper()
	v, err := NewCustomValidator()
	if err != nil {
		t.Fatalf("validator init: %v", err)
	}
	return v
}

func TestValidateStrongPassword(t *testing.T) {
	v := mustValidator(t)

	// too short
	if err := v.ValidateVar("Ab1$", "strong_password"); err == nil {
		t.Fatalf("expected error for short password")
	}
	// valid: meets 3/4 classes
	if err := v.ValidateVar("Abcd1234", "strong_password"); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
	// valid with special
	if err := v.ValidateVar("Abcd1234$", "strong_password"); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
}

func TestValidateSafeString(t *testing.T) {
	v := mustValidator(t)
	if err := v.ValidateVar("hello\nworld", "safe_string"); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
	if err := v.ValidateVar("bad\x00", "safe_string"); err == nil {
		t.Fatalf("expected error for null byte")
	}
}

func TestValidateUUID(t *testing.T) {
	v := mustValidator(t)
	if err := v.ValidateVar("550e8400-e29b-41d4-a716-446655440000", "uuid"); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
	if err := v.ValidateVar("not-a-uuid", "uuid"); err == nil {
		t.Fatalf("expected invalid uuid")
	}
}

func TestValidateEmailDomain(t *testing.T) {
	v := mustValidator(t)
	if err := v.ValidateVar("user@example.com", "email_domain"); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
	if err := v.ValidateVar("user@mailinator.com", "email_domain"); err == nil {
		t.Fatalf("expected suspicious domain rejection")
	}
}

func TestValidateNoSQLInjectionAndXSS(t *testing.T) {
	v := mustValidator(t)
	if err := v.ValidateVar("normal text", "no_sql_injection"); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
	if err := v.ValidateVar("' or 1=1--", "no_sql_injection"); err == nil {
		t.Fatalf("expected sql injection detection")
	}
	if err := v.ValidateVar("hello", "no_xss"); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
	if err := v.ValidateVar("<script>alert(1)</script>", "no_xss"); err == nil {
		t.Fatalf("expected xss detection")
	}
}

func TestValidateStructAndErrors(t *testing.T) {
	v := mustValidator(t)
	s := sampleStruct{
		Password: "weak",                // invalid
		Safe:     "ok",                  // valid
		UUID:     "not-a-uuid",          // invalid
		Email:    "user@mailinator.com", // invalid domain
		SQL:      "' or 1=1--",          // invalid
		XSS:      "<script>",            // invalid
	}
	if err := v.Validate(s); err == nil {
		t.Fatalf("expected validation error")
	} else {
		errs := v.GetValidationErrors(err)
		if len(errs) == 0 {
			t.Fatalf("expected formatted errors")
		}
	}
}
