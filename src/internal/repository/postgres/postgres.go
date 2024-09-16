package postgres

import (
	"context"
	"fmt"
	"git.codenrock.com/avito/internal/domain/dto"
	"git.codenrock.com/avito/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

const (
	DecisionApproved = "Approved"
	DecisionRejected = "Rejected"
)

func New(conn string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) GetTenders(ctx context.Context, serviceTypes []string, limit, offset int) ([]dto.TenderResponseDTO, error) {
	const op = "storage.postgres.GetTenders"

	query := `SELECT id, name, description, status, service_type, version, created_at 
              FROM tenders WHERE status = 'PUBLISHED'`

	var args []interface{}
	argIndex := 1

	if len(serviceTypes) > 0 {
		query += ` WHERE service_type = ANY($` + fmt.Sprint(argIndex) + `)`
		args = append(args, serviceTypes)
		argIndex++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(argIndex) + ` OFFSET $` + fmt.Sprint(argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var tenders []dto.TenderResponseDTO
	for rows.Next() {
		var tender dto.TenderResponseDTO
		if err := rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.Status, &tender.ServiceType, &tender.Version, &tender.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tenders = append(tenders, tender)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tenders, nil
}

func (s *Storage) CreateTender(ctx context.Context, tender dto.TenderDTO) (dto.TenderResponseDTO, error) {
	const op = "storage.postgres.CreateTender"

	query := `INSERT INTO tenders (name, description, status, service_type, organization_id, creator_username, version, created_at, updated_at)
			  VALUES ($1, $2, 'Created', $3, $4, $5, 1, NOW(), NOW()) RETURNING id, created_at`
	var tenderID uuid.UUID
	var createdAt time.Time
	var version int
	err := s.db.QueryRow(ctx, query, tender.Name, tender.Description, tender.ServiceType, tender.Status, tender.OrganizationID, tender.CreatorUsername).Scan(&tenderID, &createdAt, &version)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	newTender := dto.TenderResponseDTO{
		ID:          tenderID,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.ServiceType,
		Version:     version,
		CreatedAt:   createdAt,
	}
	return newTender, nil
}

func (s *Storage) GetUserTenders(ctx context.Context, username string, limit, offset int) ([]dto.TenderResponseDTO, error) {
	const op = "storage.postgres.GetUserTenders"

	query := `SELECT id, name, description, status, service_type, version, created_at 
              FROM tenders WHERE creator_username = $1`

	query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, query, username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var tenders []dto.TenderResponseDTO
	for rows.Next() {
		var tender dto.TenderResponseDTO
		if err = rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.Status, &tender.ServiceType, &tender.Version, &tender.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tenders = append(tenders, tender)
	}
	return tenders, nil
}

func (s *Storage) GetTenderStatus(ctx context.Context, tenderID uuid.UUID, username string) (string, error) {
	query := `SELECT status, organization_id FROM tenders WHERE id = $1`
	var status string
	var organizationID string
	err := s.db.QueryRow(ctx, query, tenderID).Scan(&status, &organizationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("tender not found")
		}
		return "", err
	}

	if status != "PUBLISHED" {
		isResponsible, err := s.IsUserResponsibleForOrganization(ctx, username, organizationID)
		if err != nil {
			return "", err
		}

		if !isResponsible {
			return "", fmt.Errorf("user is not responsible for the organization")
		}
	}

	return status, nil
}

func (s *Storage) UpdateTenderStatus(ctx context.Context, tenderID uuid.UUID, newStatus, username string) (dto.TenderResponseDTO, error) {
	const op = "storage.postgres.UpdateTenderStatus"

	query := `UPDATE tenders SET status = $1, updated_at = NOW() WHERE id = $2 AND creator_username = $3`
	result, err := s.db.Exec(ctx, query, newStatus, tenderID, username)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if result.RowsAffected() == 0 {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrUserIsNotCreatorOrTenderWasNotFound)
	}

	query = `SELECT id, name, description, status, service_type, version, created_at FROM tenders WHERE id = $1`
	var tender dto.TenderResponseDTO
	err = s.db.QueryRow(ctx, query, tenderID).Scan(&tender.ID, &tender.Name, &tender.Description, &tender.Status, &tender.ServiceType, &tender.Version, &tender.CreatedAt)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return tender, nil
}

func (s *Storage) UpdateTenderInfo(ctx context.Context, tenderID uuid.UUID, updatedData dto.UpdateTenderDTO, username string) (dto.TenderResponseDTO, error) {
	const op = "storage.postgres.UpdateTenderInfo"

	query := `UPDATE tenders SET name = COALESCE(NULLIF($1, ''), name), 
								description = COALESCE(NULLIF($2, ''), description), 
								service_type = COALESCE(NULLIF($3, ''), service_type), 
								version = version + 1, 
								updated_at = NOW() 
			  WHERE id = $4 AND creator_username = $5 RETURNING id, name, description, service_type, status, version, created_at`

	var updatedTender dto.TenderResponseDTO
	err := s.db.QueryRow(ctx, query, updatedData.Name, updatedData.Description, updatedData.ServiceType, tenderID, username).Scan(
		&updatedTender.ID, &updatedTender.Name, &updatedTender.Description, &updatedTender.ServiceType, &updatedTender.Status, &updatedTender.Version, &updatedTender.CreatedAt)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return updatedTender, nil
}

func (s *Storage) RollbackTenderVersion(ctx context.Context, tenderID uuid.UUID, version int, username string) (dto.TenderResponseDTO, error) {
	const op = "storage.postgres.RollbackTenderVersion"

	var rollbackTender dto.TenderResponseDTO
	query := `SELECT name, description, service_type, status FROM tender_versions WHERE tender_id = $1 AND version = $2`
	err := s.db.QueryRow(ctx, query, tenderID, version).Scan(&rollbackTender.Name, &rollbackTender.Description, &rollbackTender.ServiceType, &rollbackTender.Status)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	updateQuery := `UPDATE tenders SET name = $1, description = $2, service_type = $3, status = $4, version = version + 1, updated_at = NOW() 
					WHERE id = $5 AND creator_username = $6 RETURNING id, name, description, service_type, status, version, created_at`

	err = s.db.QueryRow(ctx, updateQuery, rollbackTender.Name, rollbackTender.Description, rollbackTender.ServiceType, rollbackTender.Status, tenderID, username).Scan(
		&rollbackTender.ID, &rollbackTender.Name, &rollbackTender.Description, &rollbackTender.ServiceType, &rollbackTender.Status, &rollbackTender.Version, &rollbackTender.CreatedAt)
	if err != nil {
		return dto.TenderResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return rollbackTender, nil
}

// Bids methods ---------------------------------------------------------------

func (s *Storage) CreateBid(ctx context.Context, bid *dto.BidDTO) (dto.BidResponseDTO, error) {
	const op = "storage.postgres.CreateBid"

	var tenderExists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenders WHERE id = $1)`, bid.TenderID).Scan(&tenderExists)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !tenderExists {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrTenderNotFound)
	}

	var organizationExists bool
	err = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization WHERE id = $1)`, bid.AuthorID).Scan(&organizationExists)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !organizationExists {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrOrganizationNotFound)
	}

	if bid.AuthorType == "User" {
		var userInOrganization bool
		err = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization_responsible 
                                                WHERE organization_id = $1 
                                                AND user_id = $2)`,
			bid.AuthorID, bid.AuthorID).Scan(&userInOrganization)
		if err != nil {
			return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
		}
		if !userInOrganization {
			return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrNoAssociationWithOrganization)
		}
	}

	var id uuid.UUID
	err = s.db.QueryRow(ctx, `INSERT INTO bids (name, description, tender_id, organization_id, author_type, author_id, status, version, created_at) 
	VALUES ($1, $2, $3, $4, $5, $6, 'CREATED', 1, NOW()) 
	RETURNING id`,
		bid.Name, bid.Description, bid.TenderID, bid.AuthorID, bid.AuthorType, bid.AuthorID).Scan(&id)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	var createdBid dto.BidResponseDTO
	err = s.db.QueryRow(ctx, `SELECT id, name, status, author_type, author_id, version, created_at 
		FROM bids WHERE id = $1`, id).Scan(
		&createdBid.ID,
		&createdBid.Name,
		&createdBid.Status,
		&createdBid.AuthorType,
		&createdBid.AuthorID,
		&createdBid.Version,
		&createdBid.CreatedAt,
	)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return createdBid, nil
}

func (s *Storage) GetBidsByUsername(ctx context.Context, username string, limit, offset int) ([]dto.BidResponseDTO, error) {
	query := `SELECT id, name, status, author_type, author_id, version, created_at FROM bids WHERE author_id = (SELECT id FROM employee WHERE username = $1) ORDER BY name ASC LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, query, username, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []dto.BidResponseDTO
	for rows.Next() {
		var bid dto.BidResponseDTO
		err = rows.Scan(&bid.ID, &bid.Name, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Version, &bid.CreatedAt)
		if err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}
	return bids, nil
}

func (s *Storage) GetTenderBids(ctx context.Context, tenderID string, limit, offset int) ([]dto.BidResponseDTO, error) {

	query := `SELECT id, name, status, author_type, author_id, version, created_at 
              FROM bids WHERE tender_id = $1 ORDER BY name ASC LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(ctx, query, tenderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []dto.BidResponseDTO
	for rows.Next() {
		var bid dto.BidResponseDTO
		if err = rows.Scan(&bid.ID, &bid.Name, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Version, &bid.CreatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}

	return bids, nil
}

func (s *Storage) GetBidStatus(ctx context.Context, bidID uuid.UUID, username string) (string, error) {
	const op = "storage.postgres.GetBidStatus"

	var status string
	err := s.db.QueryRow(ctx, `SELECT status FROM bids WHERE id = $1`, bidID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("%s: %w", op, repository.ErrBidNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var userAuthorized bool
	err = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization_responsible 
                                                WHERE organization_id = (SELECT organization_id FROM bids WHERE id = $1) 
                                                AND user_id = (SELECT id FROM employee WHERE username = $2))`,
		bidID, username).Scan(&userAuthorized)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	if !userAuthorized {
		return "", fmt.Errorf("%s: %w", op, repository.ErrNoPermission)
	}

	return status, nil
}

func (s *Storage) UpdateBidStatus(ctx context.Context, bidID uuid.UUID, status string, username string) (dto.BidResponseDTO, error) {
	const op = "repository.postgres.UpdateBidStatus"

	var exists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bids WHERE id = $1)`, bidID).Scan(&exists)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return dto.BidResponseDTO{}, repository.ErrBidNotFound
	}

	var userIsResponsible bool
	err = s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 
			FROM organization_responsible
			WHERE user_id = (SELECT id FROM employee WHERE username = $1)
			AND organization_id = (SELECT organization_id FROM bids WHERE id = $2)
		)`, username, bidID).Scan(&userIsResponsible)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !userIsResponsible {
		return dto.BidResponseDTO{}, repository.ErrNoPermission
	}

	_, err = s.db.Exec(ctx, `UPDATE bids SET status = $1 WHERE id = $2`, status, bidID)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	var Bid dto.BidResponseDTO
	err = s.db.QueryRow(ctx, `SELECT id, name, status, author_type, author_id, version, created_at FROM bids WHERE id = $1`, bidID).Scan(
		&Bid.ID,
		&Bid.Name,
		&Bid.Status,
		&Bid.AuthorType,
		&Bid.AuthorID,
		&Bid.Version,
		&Bid.CreatedAt,
	)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return Bid, nil
}

func (s *Storage) UpdateBid(ctx context.Context, bidID uuid.UUID, username string, updates dto.UpdateBidDTO) (dto.BidResponseDTO, error) {
	const op = "repository.postgres.UpdateBid"

	var exists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bids WHERE id = $1)`, bidID).Scan(&exists)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return dto.BidResponseDTO{}, repository.ErrBidNotFound
	}

	var userIsResponsible bool
	err = s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 
			FROM organization_responsible
			WHERE user_id = (SELECT id FROM employee WHERE username = $1)
			AND organization_id = (SELECT organization_id FROM bids WHERE id = $2)
		)`, username, bidID).Scan(&userIsResponsible)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !userIsResponsible {
		return dto.BidResponseDTO{}, repository.ErrNoPermission
	}

	_, err = s.db.Exec(ctx, `
		UPDATE bids
		SET name = COALESCE(NULLIF($1, ''), name),
			description = COALESCE(NULLIF($2, ''), description),
			version = version + 1
		WHERE id = $3
	`, updates.Name, updates.Description, bidID)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	var updatedBid dto.BidResponseDTO
	err = s.db.QueryRow(ctx, `SELECT id, name, status, author_type, author_id, version, created_at FROM bids WHERE id = $1`, bidID).Scan(
		&updatedBid.ID,
		&updatedBid.Name,
		&updatedBid.Status,
		&updatedBid.AuthorType,
		&updatedBid.AuthorID,
		&updatedBid.Version,
		&updatedBid.CreatedAt,
	)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return updatedBid, nil
}

func (s *Storage) SubmitDecision(ctx context.Context, bidID uuid.UUID, decision, username string) (dto.BidResponseDTO, error) {
	const op = "repository.postgres.SubmitDecision"

	var exists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bids WHERE id = $1)`, bidID).Scan(&exists)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrBidNotFound)
	}

	var userIsResponsible bool
	err = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization_responsible WHERE user_id = (SELECT id FROM employee WHERE username = $1))`, username).Scan(&userIsResponsible)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !userIsResponsible {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrNoPermission)
	}

	_, err = s.db.Exec(ctx, `UPDATE bids SET status = $1 WHERE id = $2`, decision, bidID)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	if decision == DecisionApproved {
		tenderID := s.getTenderIDForBid(ctx, bidID)
		if err = s.closeTender(ctx, tenderID); err != nil {
			return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	var bid dto.BidResponseDTO
	query := `SELECT id, name, status, author_type, author_id, version, created_at FROM bids WHERE id = $1`
	err = s.db.QueryRow(ctx, query, bidID).Scan(
		&bid.ID,
		&bid.Name,
		&bid.Status,
		&bid.AuthorType,
		&bid.AuthorID,
		&bid.Version,
		&bid.CreatedAt,
	)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return bid, nil
}

func (s *Storage) SendFeedback(ctx context.Context, bidID uuid.UUID, feedback, username string) (dto.BidResponseDTO, error) {
	const op = "repository.postgres.SubmitFeedback"

	var exists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bids WHERE id = $1)`, bidID).Scan(&exists)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrBidNotFound)
	}

	var userIsAuthorized bool
	err = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization_responsible WHERE user_id = (SELECT id FROM employee WHERE username = $1))`, username).Scan(&userIsAuthorized)
	if err != nil || !userIsAuthorized {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrNoPermission)
	}

	_, err = s.db.Exec(ctx, `INSERT INTO bid_feedback (bid_id, feedback, author_id) VALUES ($1, $2, (SELECT id FROM employee WHERE username = $3))`, bidID, feedback, username)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	var bid dto.BidResponseDTO
	query := `SELECT id, name, status, author_type, author_id, version, created_at FROM bids WHERE id = $1`
	err = s.db.QueryRow(ctx, query, bidID).Scan(
		&bid.ID,
		&bid.Name,
		&bid.Status,
		&bid.AuthorType,
		&bid.AuthorID,
		&bid.Version,
		&bid.CreatedAt,
	)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return bid, nil
}

func (s *Storage) RollbackBidVersion(ctx context.Context, bidID uuid.UUID, version int, username string) (dto.BidResponseDTO, error) {
	const op = "repository.postgres.RollbackBidVersion"

	var exists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bids WHERE id = $1)`, bidID).Scan(&exists)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrBidNotFound)
	}

	var versionExists bool
	err = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bid_versions WHERE bid_id = $1 AND version = $2)`, bidID, version).Scan(&versionExists)
	if err != nil || !versionExists {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrVersionNotFound)
	}

	var userIsAuthorized bool
	err = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization_responsible WHERE user_id = (SELECT id FROM employee WHERE username = $1))`, username).Scan(&userIsAuthorized)
	if err != nil || !userIsAuthorized {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, repository.ErrNoPermission)
	}

	var bid dto.UpdateBidDTO
	err = s.db.QueryRow(ctx, `SELECT name, description FROM bid_versions WHERE bid_id = $1 AND version = $2`, bidID, version).Scan(
		&bid.Name,
		&bid.Description,
	)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.Exec(ctx, `UPDATE bids SET name = $1, description = $2, version = version + 1 WHERE id = $3`,
		bid.Name, bid.Description, bidID)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	var bidVersion dto.BidResponseDTO
	err = s.db.QueryRow(ctx, `SELECT id, name, status, author_type, author_id, version, created_at FROM bids WHERE id = $1`, bidID).Scan(
		&bidVersion.ID,
		&bidVersion.Name,
		&bidVersion.Status,
		&bidVersion.AuthorType,
		&bidVersion.AuthorID,
		&bidVersion.Version,
		&bidVersion.CreatedAt,
	)
	if err != nil {
		return dto.BidResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return bidVersion, nil
}

func (s *Storage) GetBidReviews(ctx context.Context, tenderID uuid.UUID, authorUsername, requesterUsername string, limit, offset int) ([]dto.BidReviewDTO, error) {
	const op = "repository.postgres.GetBidReviews"

	var exists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenders WHERE id = $1)`, tenderID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, repository.ErrTenderNotFound)
	}

	var userIsResponsible bool
	err = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization_responsible WHERE user_id = (SELECT id FROM employee WHERE username = $1))`, requesterUsername).Scan(&userIsResponsible)
	if err != nil || !userIsResponsible {
		return nil, fmt.Errorf("%s: %w", op, repository.ErrNoPermission)
	}

	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.description, r.created_at 
		FROM bid_reviews r
		JOIN bids b ON r.bid_id = b.id
		JOIN employees e ON b.author_id = e.id
		WHERE b.tender_id = $1 AND e.username = $2
		LIMIT $3 OFFSET $4`, tenderID, authorUsername, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var reviews []dto.BidReviewDTO
	for rows.Next() {
		var review dto.BidReviewDTO
		err = rows.Scan(&review.ID, &review.Description, &review.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		reviews = append(reviews, review)
	}

	if len(reviews) == 0 {
		return nil, fmt.Errorf("%s: %w", op, repository.ErrReviewsNotFound)
	}

	return reviews, nil
}

func (s *Storage) Close() {
	s.db.Close()

	return
}

func (s *Storage) isUserResponsibleForAnyOrganization(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organization_responsibles WHERE username = $1)`
	var isResponsible bool
	err := s.db.QueryRow(ctx, query, username).Scan(&isResponsible)
	return isResponsible, err
}

func (s *Storage) IsUserResponsibleForOrganization(ctx context.Context, username, organizationID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organization_responsibles WHERE username = $1 AND organization_id = $2)`
	var isResponsible bool
	err := s.db.QueryRow(ctx, query, username, organizationID).Scan(&isResponsible)
	return isResponsible, err
}

func (s *Storage) getTenderIDForBid(ctx context.Context, bidID uuid.UUID) uuid.UUID {
	// Реализация получения ID тендера по предложению
	var tenderID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT tender_id FROM bids WHERE id = $1`, bidID).Scan(&tenderID)
	if err != nil {
		return uuid.Nil
	}
	return tenderID
}

func (s *Storage) closeTender(ctx context.Context, tenderID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE tenders SET status = 'CLOSED' WHERE id = $1`, tenderID)
	return err
}
