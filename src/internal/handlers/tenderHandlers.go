package handlers

import (
	"context"
	"git.codenrock.com/avito/internal/domain/dto"
	"git.codenrock.com/avito/internal/domain/models"
	"git.codenrock.com/avito/internal/services"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strconv"
)

type TenderService interface {
	GetTenders(ctx context.Context, serviceTypes []string, limit, offset int) ([]dto.TenderResponseDTO, error)
	CreateTender(ctx context.Context, tender *models.Tender) error
	GetUserTenders(ctx context.Context, username string, limit, offset int) ([]dto.TenderResponseDTO, error)
	GetTenderStatus(ctx context.Context, tenderID, username string) (string, error)
	UpdateTenderStatus(ctx context.Context, tenderID, newStatus, username string) (dto.TenderResponseDTO, error)
	UpdateTenderInfo(ctx context.Context, tenderID string, updatedData dto.UpdateTenderDTO, username string) (dto.TenderResponseDTO, error)
	RollbackTenderVersion(ctx context.Context, tenderID string, version int, username string) (dto.TenderResponseDTO, error)
}

type TenderHandler struct {
	log           *slog.Logger
	tenderService *services.TenderService
}

func NewTenderHandler(log *slog.Logger, tenderService *services.TenderService) *TenderHandler {
	return &TenderHandler{
		log:           log,
		tenderService: tenderService,
	}
}

func (h *TenderHandler) GetTenders(c *gin.Context) {
	limit := c.DefaultQuery("limit", "5")
	offset := c.DefaultQuery("offset", "0")
	serviceTypes := c.QueryArray("service_type")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	tenders, err := h.tenderService.GetTenders(c.Request.Context(), serviceTypes, limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tenders)
}

func (h *TenderHandler) CreateTender(c *gin.Context) {
	var newTender *models.Tender
	if err := c.BindJSON(&newTender); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	tender, err := h.tenderService.CreateTender(c.Request.Context(), newTender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, tender)
}

func (h *TenderHandler) GetTendersByUsername(c *gin.Context) {
	username := c.Query("username")
	limit := c.DefaultQuery("limit", "5")
	offset := c.DefaultQuery("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	tenders, err := h.tenderService.GetUserTenders(c.Request.Context(), username, limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tenders)
}

func (h *TenderHandler) GetTenderStatus(c *gin.Context) {
	tenderID := c.Param("tenderId")
	username := c.Query("username")
	status, err := h.tenderService.GetTenderStatus(c.Request.Context(), tenderID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *TenderHandler) UpdateTenderStatus(c *gin.Context) {
	tenderID := c.Param("tenderId")
	newStatus := c.Query("status")
	username := c.Query("username")

	tender, err := h.tenderService.UpdateTenderStatus(c.Request.Context(), tenderID, newStatus, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tender)
}

func (h *TenderHandler) UpdateTenderInfo(c *gin.Context) {
	tenderID := c.Param("tenderId")
	username := c.Query("username")
	var updatedData dto.UpdateTenderDTO
	if err := c.BindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	updatedTender, err := h.tenderService.UpdateTenderInfo(c.Request.Context(), tenderID, updatedData, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedTender)
}

func (h *TenderHandler) RollbackTenderVersion(c *gin.Context) {
	tenderID := c.Param("tenderId")
	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный номер версии"})
		return
	}
	username := c.Query("username")

	rolledBackTender, err := h.tenderService.RollbackTenderVersion(c.Request.Context(), tenderID, version, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rolledBackTender)
}
