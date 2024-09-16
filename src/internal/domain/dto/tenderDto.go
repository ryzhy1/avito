package dto

import (
	"github.com/google/uuid"
	"time"
)

type TenderDTO struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ServiceType     string    `json:"service_type"`
	Status          string    `json:"status"`
	OrganizationID  uuid.UUID `json:"organization_id"`
	CreatorUsername string    `json:"creator_username"`
}

type UpdateTenderDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ServiceType string    `json:"service_type"`
}

type TenderResponseDTO struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	ServiceType     string    `json:"service_type"`
	OrganizationID  uuid.UUID `json:"organization_id"`
	CreatorUsername string    `json:"creator_username"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Version         int       `json:"version"`
}
