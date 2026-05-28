package internal

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	RxValidateEmail = regexp.MustCompile(".+@.+\\..+")
)

type FormFieldType int

const (
	FFName FormFieldType = 1 << iota
	FFEmail
	FFPassword
	FFPhone
	FFTrade
	FFRate
	FFReference
	FFJobName
	FFSiteID
	FFStartDatetime
	FFHeadcount
	FFJobID
	FFStartTime
	FFNotes
	FFOperativeID
	FFArrivalTime
	FFResourceID
	FFQuantity
	FFDurationHours
	FFCostUnit
	FFResourceType
	FFAddress
)

func (set FormFieldType) Has(flag FormFieldType) bool {
	return set&flag == flag
}

type FormFields struct {
	Errors map[string]string
	Values map[string]string
}

func NewFormFields() *FormFields {
	return &FormFields{
		Errors: map[string]string{},
		Values: map[string]string{},
	}
}

func (ff *FormFields) ValidateFormFields(ffFlag FormFieldType) *FormFields {
	if ffFlag.Has(FFName) {
		name := strings.TrimSpace(ff.Values["name"])
		if name == "" {
			ff.Errors["name"] = "Name is required"
		} else if len(name) < 2 {
			ff.Errors["name"] = "Name must be at least 2 characters"
		}
	}

	if ffFlag.Has(FFEmail) {
		email := strings.TrimSpace(ff.Values["email"])
		if email == "" {
			ff.Errors["email"] = "Email is required"
		} else if !RxValidateEmail.MatchString(email) {
			ff.Errors["email"] = "Enter a valid email address"
		}
	}

	if ffFlag.Has(FFPassword) {
		password := ff.Values["password"]
		if password == "" {
			ff.Errors["password"] = "Password is required"
		} else if len(password) < 8 {
			ff.Errors["password"] = "Password must be at least 8 characters"
		}
	}

	if ffFlag.Has(FFPhone) {
		phone := strings.TrimSpace(ff.Values["phone"])
		if phone != "" && len(phone) < 7 {
			ff.Errors["phone"] = "Enter a valid phone number"
		}
	}

	if ffFlag.Has(FFTrade) {
		trade := strings.TrimSpace(ff.Values["trade"])
		if trade == "" {
			ff.Errors["trade"] = "Trade is required"
		}
	}

	if ffFlag.Has(FFRate) {
		rateStr := strings.TrimSpace(ff.Values["rate"])
		if rateStr == "" {
			ff.Errors["rate"] = "Rate is required"
		} else {
			rate, err := strconv.ParseUint(rateStr, 10, 64)
			if err != nil || rate == 0 {
				ff.Errors["rate"] = "Rate must be a positive whole number in pence"
			}
		}
	}

	if ffFlag.Has(FFReference) {
		ref := strings.TrimSpace(ff.Values["reference"])
		if ref == "" {
			ff.Errors["reference"] = "Job reference is required"
		}
	}

	if ffFlag.Has(FFJobName) {
		name := strings.TrimSpace(ff.Values["name"])
		if name == "" {
			ff.Errors["name"] = "Job name is required"
		}
	}

	if ffFlag.Has(FFSiteID) {
		siteID := strings.TrimSpace(ff.Values["site_id"])
		if siteID == "" {
			ff.Errors["site_id"] = "Site is required"
		}
	}

	if ffFlag.Has(FFStartDatetime) {
		start := strings.TrimSpace(ff.Values["start_datetime"])
		if start == "" {
			ff.Errors["start_datetime"] = "Start date is required"
		}
	}

	if ffFlag.Has(FFHeadcount) {
		hStr := strings.TrimSpace(ff.Values["expected_headcount"])
		if hStr == "" {
			ff.Errors["expected_headcount"] = "Headcount is required"
		} else {
			h, err := strconv.ParseUint(hStr, 10, 32)
			if err != nil || h == 0 {
				ff.Errors["expected_headcount"] = "Headcount must be at least 1"
			}
		}
	}

	if ffFlag.Has(FFJobID) {
		jobID := strings.TrimSpace(ff.Values["job_id"])
		if jobID == "" {
			ff.Errors["job_id"] = "Job is required"
		}
	}

	if ffFlag.Has(FFStartTime) {
		startTime := strings.TrimSpace(ff.Values["start_time"])
		if startTime == "" {
			ff.Errors["start_time"] = "Start time is required"
		}
	}

	if ffFlag.Has(FFNotes) {
		// notes are optional — no validation required
	}

	if ffFlag.Has(FFOperativeID) {
		opID := strings.TrimSpace(ff.Values["operative_id"])
		if opID == "" {
			ff.Errors["operative_id"] = "Operative is required"
		}
	}

	if ffFlag.Has(FFArrivalTime) {
		arrivalTime := strings.TrimSpace(ff.Values["arrival_time"])
		if arrivalTime == "" {
			ff.Errors["arrival_time"] = "Arrival time is required"
		}
	}

	if ffFlag.Has(FFResourceID) {
		resourceID := strings.TrimSpace(ff.Values["resource_id"])
		if resourceID == "" {
			ff.Errors["resource_id"] = "Resource is required"
		}
	}

	if ffFlag.Has(FFQuantity) {
		qStr := strings.TrimSpace(ff.Values["quantity"])
		if qStr == "" {
			ff.Errors["quantity"] = "Quantity is required"
		} else {
			q, err := strconv.ParseFloat(qStr, 64)
			if err != nil || q <= 0 {
				ff.Errors["quantity"] = "Quantity must be greater than zero"
			}
		}
	}

	if ffFlag.Has(FFDurationHours) {
		dStr := strings.TrimSpace(ff.Values["duration_hours"])
		if dStr == "" {
			ff.Errors["duration_hours"] = "Duration is required"
		} else {
			d, err := strconv.ParseFloat(dStr, 64)
			if err != nil || d <= 0 {
				ff.Errors["duration_hours"] = "Duration must be greater than zero"
			}
		}
	}

	if ffFlag.Has(FFCostUnit) {
		costUnit := strings.TrimSpace(ff.Values["cost_unit"])
		if costUnit == "" {
			ff.Errors["cost_unit"] = "Cost unit is required"
		} else if costUnit != "per_hour" && costUnit != "per_unit" {
			ff.Errors["cost_unit"] = "Cost unit must be per_hour or per_unit"
		}
	}

	if ffFlag.Has(FFResourceType) {
		resourceType := strings.TrimSpace(ff.Values["resource_type"])
		if resourceType == "" {
			ff.Errors["resource_type"] = "Resource type is required"
		} else if resourceType != "tool" && resourceType != "material" && resourceType != "mechanical" {
			ff.Errors["resource_type"] = "Resource type must be tool, material or mechanical"
		}
	}

	if ffFlag.Has(FFAddress) {
		address := strings.TrimSpace(ff.Values["address"])
		if address == "" {
			ff.Errors["address"] = "Address is required"
		}
	}

	return ff
}

func (ff *FormFields) HasErrors() bool {
	return len(ff.Errors) > 0
}
