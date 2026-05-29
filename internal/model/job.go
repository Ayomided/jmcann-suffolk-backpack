package model

import (
	"time"

	"github.com/google/uuid"
)

type ResourceType string

const (
	ResourceTypeTool       ResourceType = "tool"
	ResourceTypeMaterial   ResourceType = "material"
	ResourceTypeMechanical ResourceType = "mechanical"
)

type JobStatus string

const (
	JobStatusDraft     JobStatus = "draft"
	JobStatusActive    JobStatus = "active"
	JobStatusSubmitted JobStatus = "submitted"
	JobStatusCompleted JobStatus = "completed"
	JobStatusApproved  JobStatus = "approved"
)

type Resource struct {
	ID            uuid.UUID
	Name          string
	ResourceType  ResourceType
	UnitOfMeasure *string
}

type ResourceRate struct {
	ID            uuid.UUID
	ResourceID    uuid.UUID
	Rate          Money
	CostUnit      CostUnit
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}

type Job struct {
	ID            uuid.UUID
	Reference     string
	Name          string
	SiteID        uuid.UUID
	CreatedBy     uuid.UUID
	Status        JobStatus
	StartDatetime time.Time
	EndDatetime   *time.Time
	TotalCost     *Money
}

type JobSession struct {
	ID          uuid.UUID
	JobID       uuid.UUID
	SessionDate time.Time
	StartTime   time.Time
	EndTime     *time.Time
	SubmittedAt *time.Time
	SubmittedBy *uuid.UUID
	TotalCost   *Money
	Notes       *string
}

type JobOperativeRequirement struct {
	ID                uuid.UUID
	JobID             uuid.UUID
	OperativeID       *uuid.UUID
	ExpectedHeadcount uint32
	Notes             *string
}

type JobResourceRequirement struct {
	ID                    uuid.UUID
	JobID                 uuid.UUID
	ResourceID            uuid.UUID
	ExpectedQuantity      *float64
	ExpectedDurationHours *float64
	Notes                 *string
}

type JobOperative struct {
	ID             uuid.UUID
	JobID          uuid.UUID
	SessionID      uuid.UUID
	OperativeID    uuid.UUID
	ArrivalTime    time.Time
	OperativeName  *string
	DepartureTime  *time.Time
	RateSnapshot   Money
	CalculatedCost *Money
}

func (jo JobOperative) HoursWorked() *float64 {
	if jo.DepartureTime == nil {
		return nil
	}
	hours := jo.DepartureTime.Sub(jo.ArrivalTime).Hours()
	return &hours
}

func (jo JobOperative) Calculate() *Money {
	hours := jo.HoursWorked()
	if hours == nil {
		return nil
	}
	result := jo.RateSnapshot.Multiply(*hours)
	return &result
}

type JobResource struct {
	ID             uuid.UUID
	JobID          uuid.UUID
	SessionID      uuid.UUID
	ResourceID     uuid.UUID
	ResourceName   string
	ResourceType   ResourceType
	Quantity       *float64
	DurationHours  *float64
	ArrivalTime    *time.Time
	RateSnapshot   Money
	CalculatedCost *Money
}

func (jr JobResource) Calculate() Money {
	switch {
	case jr.Quantity != nil:
		return jr.RateSnapshot.Multiply(*jr.Quantity)
	case jr.DurationHours != nil:
		return jr.RateSnapshot.Multiply(*jr.DurationHours)
	default:
		return NewMoney("GBP", 0)
	}
}

type CostCalculation struct {
	ID                  uuid.UUID
	JobID               uuid.UUID
	OperativeCostsTotal Money
	ResourceCostsTotal  Money
	GrandTotal          Money
	CalculatedAt        time.Time
	CalculatedBy        uuid.UUID
}

func NewCostCalculation(
	jobID uuid.UUID,
	operativeCosts []Money,
	resourceCosts []Money,
	calculatedBy uuid.UUID,
) (CostCalculation, error) {
	opTotal := NewMoney("GBP", 0)
	for _, c := range operativeCosts {
		result, err := opTotal.Add(c)
		if err != nil {
			return CostCalculation{}, err
		}
		opTotal = result
	}

	resTotal := NewMoney("GBP", 0)
	for _, c := range resourceCosts {
		result, err := resTotal.Add(c)
		if err != nil {
			return CostCalculation{}, err
		}
		resTotal = result
	}

	grandTotal, err := opTotal.Add(resTotal)
	if err != nil {
		return CostCalculation{}, err
	}

	return CostCalculation{
		ID:                  uuid.New(),
		JobID:               jobID,
		OperativeCostsTotal: opTotal,
		ResourceCostsTotal:  resTotal,
		GrandTotal:          grandTotal,
		CalculatedAt:        time.Now().UTC(),
		CalculatedBy:        calculatedBy,
	}, nil
}
