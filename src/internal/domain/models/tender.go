package models

import "github.com/google/uuid"

type Tender struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ServiceType     string    `json:"service_type"`
	Status          string    `json:"status"`
	OrganizationID  uuid.UUID `json:"organization_id"`
	CreatorUsername string    `json:"creator_username"`
}
