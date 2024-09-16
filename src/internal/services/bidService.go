package services

import (
	"context"
	"fmt"
	"git.codenrock.com/avito/internal/domain/dto"
	"github.com/google/uuid"
	"log/slog"
)

type BidStorage interface {
	CreateBid(ctx context.Context, bid *dto.BidDTO) (dto.BidResponseDTO, error)
	GetBidsByUsername(ctx context.Context, username string, limit, offset int) ([]dto.BidResponseDTO, error)
	GetTenderBids(ctx context.Context, tenderID string, limit, offset int) ([]dto.BidResponseDTO, error)
	GetBidStatus(ctx context.Context, bidID uuid.UUID, username string) (string, error)
	UpdateBid(ctx context.Context, bidID uuid.UUID, username string, updates dto.UpdateBidDTO) (dto.BidResponseDTO, error)
	UpdateBidStatus(ctx context.Context, bidID uuid.UUID, status string, username string) (dto.BidResponseDTO, error)
	SubmitDecision(ctx context.Context, bidID uuid.UUID, decision, username string) (dto.BidResponseDTO, error)
	SendFeedback(ctx context.Context, bidID uuid.UUID, feedback, username string) (dto.BidResponseDTO, error)
	RollbackBidVersion(ctx context.Context, bidID uuid.UUID, version int, username string) (dto.BidResponseDTO, error)
	GetBidReviews(ctx context.Context, tenderID uuid.UUID, authorUsername, requesterUsername string, limit, offset int) ([]dto.BidReviewDTO, error)
}

type BidService struct {
	log *slog.Logger
	db  BidStorage
}

var (
	ErrUsernameFieldEmpty = fmt.Errorf("username field is empty")
)

func NewBidService(log *slog.Logger, db BidStorage) *BidService {
	return &BidService{
		log: log,
		db:  db,
	}
}

func (s *BidService) CreateBid(ctx context.Context, bid *dto.BidDTO) (dto.BidResponseDTO, error) {
	const op = "services.bidService.CreateBid"

	log := s.log.With(
		slog.String("op", op),
		slog.String("username", bid.Name),
	)

	if bid.Name == "" {
		log.Error("username is empty")
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, ErrUsernameFieldEmpty)
	}

	log.Info("Creating bid")

	bidResponse, err := s.db.CreateBid(ctx, bid)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Created bid")

	return bidResponse, nil
}

func (s *BidService) GetUserBids(ctx context.Context, username string, limit, offset int) ([]dto.BidResponseDTO, error) {
	const op = "services.bidService.GetUserBids"

	log := s.log.With(
		slog.String("op", op),
		slog.String("username", username),
	)

	log.Info("Getting user bids")

	bids, err := s.db.GetBidsByUsername(ctx, username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Got user bids")

	return bids, nil
}

func (s *BidService) GetTenderBids(ctx context.Context, tenderID, username string, limit, offset int) ([]dto.BidResponseDTO, error) {
	const op = "services.bidService.GetTenderBids"

	log := s.log.With(
		slog.String("op", op),
		slog.String("tenderID", tenderID),
	)

	if tenderID == "" {
		log.Error("tenderID is empty")
		return nil, fmt.Errorf("%s: %w", op, ErrTenderIDFieldEmpty)
	}

	log.Info("Getting tender bids")

	bids, err := s.db.GetTenderBids(ctx, tenderID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Got tender bids")

	return bids, nil
}

func (s *BidService) GetBidStatus(ctx context.Context, bidID string, username string) (string, error) {
	const op = "services.bidService.GetBidStatus"

	log := s.log.With(
		slog.String("op", op),
		slog.String("bidID", bidID),
	)

	bidUUID, err := uuid.Parse(bidID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Getting bid status")

	status, err := s.db.GetBidStatus(ctx, bidUUID, username)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Got bid status", slog.String("status", status))

	return status, nil
}

func (s *BidService) UpdateBid(ctx context.Context, bidID string, username string, updates dto.UpdateBidDTO) (dto.BidResponseDTO, error) {
	const op = "services.bidService.UpdateBid"

	log := s.log.With(
		slog.String("op", op),
		slog.String("bidID", bidID),
	)

	bidUUID, err := uuid.Parse(bidID)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Updating bid")

	bidResponse, err := s.db.UpdateBid(ctx, bidUUID, username, updates)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Bid updated")

	return bidResponse, nil
}

func (s *BidService) UpdateBidStatus(ctx context.Context, bidID string, status string, username string) (dto.BidResponseDTO, error) {
	const op = "services.bidService.UpdateBidStatus"

	log := s.log.With(
		slog.String("op", op),
		slog.String("bidID", bidID),
	)

	bidUUID, err := uuid.Parse(bidID)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Updating bid status")

	bidResponse, err := s.db.UpdateBidStatus(ctx, bidUUID, status, username)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Bid status updated")

	return bidResponse, nil
}

func (s *BidService) SubmitDecision(ctx context.Context, bidID string, decision string, username string) (dto.BidResponseDTO, error) {
	const op = "services.bidService.SubmitDecision"

	log := s.log.With(
		slog.String("op", op),
		slog.String("bidID", bidID),
	)

	bidUUID, err := uuid.Parse(bidID)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Submitting decision")

	bidResponse, err := s.db.SubmitDecision(ctx, bidUUID, decision, username)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Decision submitted")

	return bidResponse, nil
}

func (s *BidService) SendFeedback(ctx context.Context, bidID string, feedback, username string) (dto.BidResponseDTO, error) {
	const op = "services.bidService.SendFeedback"

	log := s.log.With(
		slog.String("op", op),
		slog.String("bidID", bidID),
	)

	bidUUID, err := uuid.Parse(bidID)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Sending feedback")

	bidResponse, err := s.db.SendFeedback(ctx, bidUUID, feedback, username)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Feedback sent")

	return bidResponse, nil
}

func (s *BidService) RollbackBidVersion(ctx context.Context, bidID string, version int, username string) (dto.BidResponseDTO, error) {
	const op = "services.bidService.RollbackBidVersion"

	log := s.log.With(
		slog.String("op", op),
		slog.String("bidID", bidID),
	)

	bidUUID, err := uuid.Parse(bidID)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Rolling back bid version")

	bidResponse, err := s.db.RollbackBidVersion(ctx, bidUUID, version, username)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Bid version rolled back")

	return bidResponse, nil
}

func (s *BidService) GetBidReviews(ctx context.Context, tenderID, authorUsername, requesterUsername string, limit, offset int) ([]dto.BidReviewDTO, error) {
	const op = "services.bidService.GetBidReviews"

	log := s.log.With(
		slog.String("op", op),
		slog.String("tenderID", tenderID),
	)

	tenderUUID, err := uuid.Parse(tenderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Getting bid reviews")

	reviews, err := s.db.GetBidReviews(ctx, tenderUUID, authorUsername, requesterUsername, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Got bid reviews")

	return reviews, nil
}
