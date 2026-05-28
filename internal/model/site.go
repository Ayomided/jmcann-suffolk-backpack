package model

import "github.com/google/uuid"

type Site struct {
	ID      uuid.UUID
	Name    string
	Address string
}
