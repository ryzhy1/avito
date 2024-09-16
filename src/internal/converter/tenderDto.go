package converter

import (
	"git.codenrock.com/avito/internal/domain/dto"
	"git.codenrock.com/avito/internal/domain/models"
)

func ToCreateTenderDTO(tender *models.Tender) dto.TenderDTO {
	return dto.TenderDTO{
		Name:            tender.Name,
		Description:     tender.Description,
		ServiceType:     tender.ServiceType,
		Status:          tender.Status,
		OrganizationID:  tender.OrganizationID,
		CreatorUsername: tender.CreatorUsername,
	}
}
