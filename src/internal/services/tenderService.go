package services

import (
	"context"
	"fmt"
	"git.codenrock.com/avito/internal/converter"
	"git.codenrock.com/avito/internal/domain/dto"
	"git.codenrock.com/avito/internal/domain/models"
	"github.com/google/uuid"
	"log/slog"
)

type Storage interface {
	GetTenders(ctx context.Context, serviceTypes []string, limit, offset int) ([]dto.TenderResponseDTO, error)
	CreateTender(ctx context.Context, tender dto.TenderDTO) (dto.TenderResponseDTO, error)
	GetUserTenders(ctx context.Context, username string, limit, offset int) ([]dto.TenderResponseDTO, error)
	GetTenderStatus(ctx context.Context, tenderID uuid.UUID, username string) (string, error)
	UpdateTenderStatus(ctx context.Context, tenderID uuid.UUID, newStatus, username string) (dto.TenderResponseDTO, error)
	UpdateTenderInfo(ctx context.Context, tenderID uuid.UUID, updatedData dto.UpdateTenderDTO, username string) (dto.TenderResponseDTO, error)
	RollbackTenderVersion(ctx context.Context, tenderID uuid.UUID, version int, username string) (dto.TenderResponseDTO, error)
	IsUserResponsibleForOrganization(ctx context.Context, username, organizationID string) (bool, error)
}

type TenderService struct {
	log *slog.Logger
	db  Storage
}

var (
	ErrTenderIDFieldEmpty = fmt.Errorf("tender id field is empty")
)

func NewTenderService(log *slog.Logger, db Storage) *TenderService {
	return &TenderService{
		log: log,
		db:  db,
	}
}

func (s *TenderService) GetTenders(ctx context.Context, serviceTypes []string, limit, offset int) ([]dto.TenderResponseDTO, error) {
	const op = "services.tenderService.GetTenders"

	s.log.Info("Get tenders", slog.String("op", op))

	tenders, err := s.db.GetTenders(ctx, serviceTypes, limit, offset)
	if err != nil {
		s.log.Error("failed to hash password", err)

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tenders, nil
}

func (s *TenderService) CreateTender(ctx context.Context, tender *models.Tender) (dto.TenderResponseDTO, error) {
	const op = "services.tenderService.CreateTender"

	s.log.With(
		slog.String("op", op),
		slog.String("name", tender.Name),
	)

	tenderDto := converter.ToCreateTenderDTO(tender)

	isResponsible, err := s.db.IsUserResponsibleForOrganization(ctx, tender.CreatorUsername, tender.OrganizationID.String())
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	if !isResponsible {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	createdTender, err := s.db.CreateTender(ctx, tenderDto)

	if err != nil {
		s.log.Error("failed to hash password", err)

		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return createdTender, nil
}

func (s *TenderService) GetUserTenders(ctx context.Context, username string, limitInt, offsetInt int) ([]dto.TenderResponseDTO, error) {
	const op = "services.tenderService.GetUserTenders"

	log := s.log.With(
		slog.String("op", op),
		slog.String("username", username),
	)

	log.Info("Getting user tenders", slog.Int("limit", limitInt), slog.Int("offset", offsetInt))

	tenders, err := s.db.GetUserTenders(ctx, username, limitInt, offsetInt)
	if tenders == nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tenders, nil
}

func (s *TenderService) GetTenderStatus(ctx context.Context, tenderID, username string) (string, error) {
	const op = "services.tenderService.GetTenderStatus"

	log := s.log.With(
		slog.String("op", op),
		slog.String("tenderID", tenderID),
	)

	if tenderID == "" {
		return "", fmt.Errorf("%s: %w", op, ErrTenderIDFieldEmpty)
	}

	tenderUUID, err := uuid.Parse(tenderID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Getting tender status")

	status, err := s.db.GetTenderStatus(ctx, tenderUUID, username)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Got tender status", slog.String("status", status))

	return status, nil
}

func (s *TenderService) UpdateTenderStatus(ctx context.Context, tenderID, newStatus, username string) (dto.TenderResponseDTO, error) {
	const op = "services.tenderService.UpdateTenderStatus"

	log := s.log.With(
		slog.String("op", op),
		slog.String("tenderID", tenderID),
	)

	if tenderID == "" {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, ErrTenderIDFieldEmpty)
	}

	tenderUUID, err := uuid.Parse(tenderID)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Updating tender status")

	tender, err := s.db.UpdateTenderStatus(ctx, tenderUUID, newStatus, username)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return tender, nil
}

func (s *TenderService) UpdateTenderInfo(ctx context.Context, tenderID string, updatedData dto.UpdateTenderDTO, username string) (dto.TenderResponseDTO, error) {
	const op = "services.tenderService.UpdateTender"

	log := s.log.With(
		slog.String("op", op),
		slog.String("tenderID", tenderID),
	)

	if tenderID == "" {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, ErrTenderIDFieldEmpty)
	}

	tenderUUID, err := uuid.Parse(tenderID)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Updating tender")

	tender, err := s.db.UpdateTenderInfo(ctx, tenderUUID, updatedData, username)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Tander updated", slog.String("tenderID", tenderID))

	return tender, nil
}

func (s *TenderService) RollbackTenderVersion(ctx context.Context, tenderID string, version int, username string) (dto.TenderResponseDTO, error) {
	const op = "services.tenderService.RollbackTenderVersion"

	log := s.log.With(
		slog.String("op", op),
		slog.String("tenderID", tenderID),
	)

	if tenderID == "" {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, ErrTenderIDFieldEmpty)
	}

	tenderUUID, err := uuid.Parse(tenderID)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Rolling back tender version")

	tender, err := s.db.RollbackTenderVersion(ctx, tenderUUID, version, username)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Tender version rolled back", slog.String("tenderID", tenderID))

	return tender, nil
}
