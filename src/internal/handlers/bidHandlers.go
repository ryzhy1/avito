package handlers

import (
	"context"
	"database/sql"
	"errors"
	"git.codenrock.com/avito/internal/domain/dto"
	"git.codenrock.com/avito/internal/repository"
	"git.codenrock.com/avito/internal/services"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strconv"
)

type BidService interface {
	CreateBid(ctx context.Context, bid *dto.BidDTO) (dto.BidResponseDTO, error)
	GetUserBids(ctx context.Context, username string, limit, offset int) ([]dto.BidResponseDTO, error)
	GetTenderBids(ctx context.Context, tenderID, username string, limit, offset int) ([]dto.BidResponseDTO, error)
	GetBidStatus(ctx context.Context, bidID string, username string) (string, error)
	UpdateBid(ctx context.Context, bidID string, username string, updates dto.UpdateBidDTO) (dto.BidResponseDTO, error)
	UpdateBidStatus(ctx context.Context, bidID string, status string, username string) (dto.BidResponseDTO, error)
	SubmitDecision(ctx context.Context, bidID string, decision string, username string) (dto.BidResponseDTO, error)
	SendFeedback(ctx context.Context, bidID string, feedback, username string) (dto.BidResponseDTO, error)
	RollbackBidVersion(ctx context.Context, bidID string, version int, username string) (dto.BidResponseDTO, error)
	GetBidReviews(ctx context.Context, tenderID, authorUsername, requesterUsername string, limit, offset int) ([]dto.BidReviewDTO, error)
}

type BidHandler struct {
	log        *slog.Logger
	bidService *services.BidService
}

func NewBidHandler(log *slog.Logger, bidService *services.BidService) *BidHandler {
	return &BidHandler{
		log:        log,
		bidService: bidService,
	}
}

func (h *BidHandler) CreateBid(c *gin.Context) {
	var bidDTO *dto.BidDTO
	if err := c.ShouldBindJSON(&bidDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Неверный формат запроса"})
		return
	}

	bid, err := h.bidService.CreateBid(c.Request.Context(), bidDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Ошибка при создании предложения"})
		return
	}

	c.JSON(http.StatusOK, bid)

}

func (h *BidHandler) GetUserBids(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Username is required"})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid limit value"})
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid offset value"})
	}

	bids, err := h.bidService.GetUserBids(context.Background(), username, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bids)
}

func (h *BidHandler) GetTenderBids(c *gin.Context) {
	tenderID := c.Param("tenderId")
	if tenderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "tenderId is required"})
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "username is required"})
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid limit value"})
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid offset value"})
	}

	if tenderID == "" || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "tenderId and username are required"})
		return
	}

	bids, err := h.bidService.GetTenderBids(c, tenderID, username, limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"reason": "No bids found for this tender"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, bids)
}

func (h *BidHandler) GetBidStatus(c *gin.Context) {
	bidID := c.Param("bidId")
	username := c.Query("username")

	status, err := h.bidService.GetBidStatus(c.Request.Context(), bidID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
	}

	c.JSON(http.StatusOK, status)
}

func (h *BidHandler) UpdateBid(c *gin.Context) {
	bidID := c.Param("bidId")
	if bidID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Bid id is required"})
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Username is required"})
	}

	var updateBidDTO dto.UpdateBidDTO
	if err := c.ShouldBindJSON(&updateBidDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Ошибка при обработке данных запроса"})
	}

	updatedBid, err := h.bidService.UpdateBid(c.Request.Context(), bidID, username, updateBidDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
	}

	c.JSON(http.StatusOK, updatedBid)
}

func (h *BidHandler) UpdateBidStatus(c *gin.Context) {
	bidID := c.Param("bidId")

	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Status is required"})
		return
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Username is required"})
		return
	}

	updatedBid, err := h.bidService.UpdateBidStatus(c.Request.Context(), bidID, status, username)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrBidNotFound):
			c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		case errors.Is(err, repository.ErrNoPermission):
			c.JSON(http.StatusForbidden, gin.H{"reason": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update bid status"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedBid)
}

func (h *BidHandler) SubmitDecision(c *gin.Context) {
	bidID := c.Param("bidId")
	if bidID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Bid id is required"})
		return
	}

	decision := c.Query("decision")
	if decision == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Decision is required"})
		return
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Username is required"})
		return
	}

	if decision != "Approved" && decision != "Rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Неверное значение решения"})
		return
	}

	bidResult, err := h.bidService.SubmitDecision(c.Request.Context(), bidID, decision, username)
	if err != nil {
		switch err {
		case repository.ErrBidNotFound:
			c.JSON(http.StatusNotFound, gin.H{"reason": "Предложение не найдено"})
		case repository.ErrNoPermission:
			c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		case repository.ErrTenderCloseFailed:
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Не удалось закрыть тендер"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, bidResult)
}

func (h *BidHandler) SendFeedback(c *gin.Context) {
	bidID := c.Param("bidId")
	if bidID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Bid id is required"})
		return
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Username is required"})
		return
	}

	feedback := c.Query("bidFeedback")
	if feedback == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Feedback is required"})
		return
	}

	updatedBid, err := h.bidService.SendFeedback(c.Request.Context(), bidID, feedback, username)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrBidNotFound):
			c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		case errors.Is(err, repository.ErrNoPermission):
			c.JSON(http.StatusForbidden, gin.H{"reason": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to submit feedback"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedBid)
}

func (h *BidHandler) RollbackBidVersion(c *gin.Context) {
	bidID := c.Param("bidId")
	if bidID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Bid id is required"})
	}

	versionParam := c.Param("version")
	if versionParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Version is required"})
	}

	version, err := strconv.Atoi(versionParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid version format"})
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Username is required"})
	}

	updatedBid, err := h.bidService.RollbackBidVersion(c.Request.Context(), bidID, version, username)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrBidNotFound):
			c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		case errors.Is(err, repository.ErrVersionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"reason": "Version not found"})
		case errors.Is(err, repository.ErrNoPermission):
			c.JSON(http.StatusForbidden, gin.H{"reason": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to rollback bid version"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedBid)
}

func (h *BidHandler) GetBidReviews(c *gin.Context) {
	tenderID := c.Param("tenderId")

	authorUsername := c.Query("authorUsername")
	if authorUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Author username is required"})
	}

	requesterUsername := c.Query("requesterUsername")
	if requesterUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Requester username is required"})
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid limit value"})
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid offset value"})
	}

	reviews, err := h.bidService.GetBidReviews(c.Request.Context(), tenderID, authorUsername, requesterUsername, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNoPermission):
			c.JSON(http.StatusForbidden, gin.H{"reason": "Insufficient permissions"})
		case errors.Is(err, repository.ErrTenderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		case errors.Is(err, repository.ErrReviewsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"reason": "Reviews not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to get reviews"})
		}
		return
	}

	c.JSON(http.StatusOK, reviews)
}
