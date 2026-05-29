package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleOperative UserRole = "operative"
	UserRoleQS        UserRole = "qs"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	Role         UserRole
	PasswordHash string
}

type Operative struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	Email  string
	Phone  string
	Trade  *string
	Rate   *Money
}

type OperativeRate struct {
	ID            uuid.UUID
	OperativeID   uuid.UUID
	RatePerHour   Money
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}
