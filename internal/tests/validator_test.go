package tests

import (
	"testing"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal"
)

func TestHasNameFieldType(t *testing.T) {
	formFieldTypes := internal.FFName | internal.FFAddress | internal.FFEmail
	if formFieldTypes.Has(internal.FFName) != true {
		t.Error("expected form field types to contain Name")
	}
}

func TestHasDoesNotContainUnsetFlag(t *testing.T) {
	formFieldTypes := internal.FFName | internal.FFAddress
	if formFieldTypes.Has(internal.FFEmail) != false {
		t.Error("expected form field types not to contain Email")
	}
}

func TestValidateNameEmpty(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["name"] = ""
	ff.ValidateFormFields(internal.FFName)
	if _, ok := ff.Errors["name"]; !ok {
		t.Error("expected name error when name is empty")
	}
}

func TestValidateNameTooShort(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["name"] = "a"
	ff.ValidateFormFields(internal.FFName)
	if _, ok := ff.Errors["name"]; !ok {
		t.Error("expected name error when name is too short")
	}
}

func TestValidateNameValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["name"] = "John Hamm"
	ff.ValidateFormFields(internal.FFName)
	if _, ok := ff.Errors["name"]; ok {
		t.Error("expected no name error for valid name")
	}
}

func TestValidateEmailEmpty(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["email"] = ""
	ff.ValidateFormFields(internal.FFEmail)
	if _, ok := ff.Errors["email"]; !ok {
		t.Error("expected email error when email is empty")
	}
}

func TestValidateEmailInvalid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["email"] = "notanemail"
	ff.ValidateFormFields(internal.FFEmail)
	if _, ok := ff.Errors["email"]; !ok {
		t.Error("expected email error for invalid email")
	}
}

func TestValidateEmailValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["email"] = "john@backpack.dev"
	ff.ValidateFormFields(internal.FFEmail)
	if _, ok := ff.Errors["email"]; ok {
		t.Error("expected no email error for valid email")
	}
}

func TestValidatePasswordEmpty(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["password"] = ""
	ff.ValidateFormFields(internal.FFPassword)
	if _, ok := ff.Errors["password"]; !ok {
		t.Error("expected password error when password is empty")
	}
}

func TestValidatePasswordTooShort(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["password"] = "abc"
	ff.ValidateFormFields(internal.FFPassword)
	if _, ok := ff.Errors["password"]; !ok {
		t.Error("expected password error when password is too short")
	}
}

func TestValidatePasswordValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["password"] = "password123"
	ff.ValidateFormFields(internal.FFPassword)
	if _, ok := ff.Errors["password"]; ok {
		t.Error("expected no password error for valid password")
	}
}

func TestValidateRateEmpty(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["rate"] = ""
	ff.ValidateFormFields(internal.FFRate)
	if _, ok := ff.Errors["rate"]; !ok {
		t.Error("expected rate error when rate is empty")
	}
}

func TestValidateRateZero(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["rate"] = "0"
	ff.ValidateFormFields(internal.FFRate)
	if _, ok := ff.Errors["rate"]; !ok {
		t.Error("expected rate error when rate is zero")
	}
}

func TestValidateRateNonNumeric(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["rate"] = "abc"
	ff.ValidateFormFields(internal.FFRate)
	if _, ok := ff.Errors["rate"]; !ok {
		t.Error("expected rate error for non-numeric rate")
	}
}

func TestValidateRateValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["rate"] = "1800"
	ff.ValidateFormFields(internal.FFRate)
	if _, ok := ff.Errors["rate"]; ok {
		t.Error("expected no rate error for valid rate")
	}
}

func TestValidateSiteIDEmpty(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["site_id"] = ""
	ff.ValidateFormFields(internal.FFSiteID)
	if _, ok := ff.Errors["site_id"]; !ok {
		t.Error("expected site_id error when site is empty")
	}
}

func TestValidateSiteIDValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["site_id"] = "some-uuid"
	ff.ValidateFormFields(internal.FFSiteID)
	if _, ok := ff.Errors["site_id"]; ok {
		t.Error("expected no site_id error for valid site")
	}
}

func TestValidateHeadcountEmpty(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["expected_headcount"] = ""
	ff.ValidateFormFields(internal.FFHeadcount)
	if _, ok := ff.Errors["expected_headcount"]; !ok {
		t.Error("expected headcount error when headcount is empty")
	}
}

func TestValidateHeadcountZero(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["expected_headcount"] = "0"
	ff.ValidateFormFields(internal.FFHeadcount)
	if _, ok := ff.Errors["expected_headcount"]; !ok {
		t.Error("expected headcount error when headcount is zero")
	}
}

func TestValidateHeadcountValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["expected_headcount"] = "3"
	ff.ValidateFormFields(internal.FFHeadcount)
	if _, ok := ff.Errors["expected_headcount"]; ok {
		t.Error("expected no headcount error for valid headcount")
	}
}

func TestValidateQuantityEmpty(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["quantity"] = ""
	ff.ValidateFormFields(internal.FFQuantity)
	if _, ok := ff.Errors["quantity"]; !ok {
		t.Error("expected quantity error when quantity is empty")
	}
}

func TestValidateQuantityZero(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["quantity"] = "0"
	ff.ValidateFormFields(internal.FFQuantity)
	if _, ok := ff.Errors["quantity"]; !ok {
		t.Error("expected quantity error when quantity is zero")
	}
}

func TestValidateQuantityValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["quantity"] = "5.5"
	ff.ValidateFormFields(internal.FFQuantity)
	if _, ok := ff.Errors["quantity"]; ok {
		t.Error("expected no quantity error for valid quantity")
	}
}

func TestValidateDurationEmpty(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["duration_hours"] = ""
	ff.ValidateFormFields(internal.FFDurationHours)
	if _, ok := ff.Errors["duration_hours"]; !ok {
		t.Error("expected duration error when duration is empty")
	}
}

func TestValidateDurationValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["duration_hours"] = "4.5"
	ff.ValidateFormFields(internal.FFDurationHours)
	if _, ok := ff.Errors["duration_hours"]; ok {
		t.Error("expected no duration error for valid duration")
	}
}

func TestValidateCostUnitInvalid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["cost_unit"] = "per_day"
	ff.ValidateFormFields(internal.FFCostUnit)
	if _, ok := ff.Errors["cost_unit"]; !ok {
		t.Error("expected cost_unit error for invalid cost unit")
	}
}

func TestValidateCostUnitValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["cost_unit"] = "per_hour"
	ff.ValidateFormFields(internal.FFCostUnit)
	if _, ok := ff.Errors["cost_unit"]; ok {
		t.Error("expected no cost_unit error for valid cost unit")
	}
}

func TestValidateResourceTypeInvalid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["resource_type"] = "vehicle"
	ff.ValidateFormFields(internal.FFResourceType)
	if _, ok := ff.Errors["resource_type"]; !ok {
		t.Error("expected resource_type error for invalid type")
	}
}

func TestValidateResourceTypeValid(t *testing.T) {
	for _, rt := range []string{"tool", "material", "mechanical"} {
		ff := internal.NewFormFields()
		ff.Values["resource_type"] = rt
		ff.ValidateFormFields(internal.FFResourceType)
		if _, ok := ff.Errors["resource_type"]; ok {
			t.Errorf("expected no resource_type error for valid type: %s", rt)
		}
	}
}

func TestValidateAddressEmpty(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["address"] = ""
	ff.ValidateFormFields(internal.FFAddress)
	if _, ok := ff.Errors["address"]; !ok {
		t.Error("expected address error when address is empty")
	}
}

func TestValidateAddressValid(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["address"] = "Woodbridge Road, Ipswich"
	ff.ValidateFormFields(internal.FFAddress)
	if _, ok := ff.Errors["address"]; ok {
		t.Error("expected no address error for valid address")
	}
}

func TestHasErrorsTrue(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["email"] = ""
	ff.ValidateFormFields(internal.FFEmail)
	if !ff.HasErrors() {
		t.Error("expected HasErrors to return true when there are errors")
	}
}

func TestHasErrorsFalse(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["email"] = "john@backpack.dev"
	ff.ValidateFormFields(internal.FFEmail)
	if ff.HasErrors() {
		t.Error("expected HasErrors to return false when there are no errors")
	}
}

func TestMultipleFlagsOnlyValidatesSelected(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["email"] = "john@backpack.dev"
	ff.Values["name"] = ""
	ff.ValidateFormFields(internal.FFEmail)
	if _, ok := ff.Errors["name"]; ok {
		t.Error("expected name not to be validated when FFName flag not set")
	}
}

func TestMultipleFlagsValidatesAll(t *testing.T) {
	ff := internal.NewFormFields()
	ff.Values["email"] = ""
	ff.Values["name"] = ""
	ff.ValidateFormFields(internal.FFEmail | internal.FFName)
	if _, ok := ff.Errors["email"]; !ok {
		t.Error("expected email error")
	}
	if _, ok := ff.Errors["name"]; !ok {
		t.Error("expected name error")
	}
}
