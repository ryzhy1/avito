package models

type Bid struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	TenderID        int    `json:"tender_id"`
	OrganizationID  int    `json:"organization_id"`
	CreatorUsername string `json:"creator_username"`
}
