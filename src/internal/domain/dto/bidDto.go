package dto

import (
	"github.com/google/uuid"
	"time"
)

type BidDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TenderID    int    `json:"tender_id"`
	AuthorType  string `json:"author_type"`
	AuthorID    int    `json:"author_id"`
}

type UpdateBidDTO struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type BidResponseDTO struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	AuthorType string    `json:"author_type"`
	AuthorID   int       `json:"author_id"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
