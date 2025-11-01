package handlers

import (
	"net/http"
	"strconv"

	"github.com/Qovix/Qovix-go/internal/middleware"
	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/Qovix/Qovix-go/internal/repository"
	"github.com/Qovix/Qovix-go/internal/services"
	"github.com/Qovix/Qovix-go/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AIHandler struct {
	aiService       *services.AIService
	databaseService *services.DatabaseService
	log             *logrus.Logger
}

func NewAIHandler(aiService *services.AIService, databaseService *services.DatabaseService) *AIHandler {
	return &AIHandler{
		aiService:       aiService,
		databaseService: databaseService,
		log:             logger.GetLogger(),
	}
}

func (h *AIHandler) GenerateSQL(c *gin.Context) {
	var req SQLQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	if req.Database == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Database name is required",
		})
		return
	}

	if len(req.Tables) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "At least one table must be specified",
		})
		return
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Query is required",
		})
		return
	}

	user, exists := middleware.GetUserFromContext(c)
	if exists {
		h.log.Infof("User %s requesting SQL generation for database: %s", user.Email, req.Database)
	}

	serviceReq := services.SQLQueryRequest{
		Database:   req.Database,
		Tables:     req.Tables,
		Query:      req.Query,
		MaxRetries: req.MaxRetries,
	}

	if req.ConnectionID != "" {
		if err := h.aiService.EnhanceRequestWithSchema(c.Request.Context(), &serviceReq, h.databaseService, req.ConnectionID); err != nil {
			h.log.Warnf("Failed to enhance request with schema: %v", err)
		}
	}

	response, err := h.aiService.GenerateSQLQuery(c.Request.Context(), serviceReq)
	if err != nil {
		h.log.Errorf("Failed to generate SQL: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "generation_failed",
			Message: "Failed to generate SQL query",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	h.log.Infof("Generated SQL query for user query '%s': %s", req.Query, response.SQLQuery)

	c.JSON(http.StatusOK, response)
}

func (h *AIHandler) ValidateSQL(c *gin.Context) {
	var req ValidateSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	if req.SQLQuery == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "SQL query is required",
		})
		return
	}

	tempService := services.NewAIService("")

	isValid, errors := tempService.ValidateSQL(req.SQLQuery)

	response := ValidateSQLResponse{
		IsValid:          isValid,
		ValidationErrors: errors,
		SQLQuery:         req.SQLQuery,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AIHandler) ExplainSQL(c *gin.Context) {
	var req ExplainSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	if req.SQLQuery == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "SQL query is required",
		})
		return
	}

	explanation, err := h.aiService.ExplainSQL(c.Request.Context(), req.SQLQuery)
	if err != nil {
		h.log.Errorf("Failed to explain SQL: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "explanation_failed",
			Message: "Failed to explain SQL query",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	response := ExplainSQLResponse{
		SQLQuery:    req.SQLQuery,
		Explanation: explanation,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AIHandler) GetQueryHistory(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not found in context",
		})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	h.log.Infof("User %s requesting query history with limit %d, offset %d", user.Email, limit, offset)

	response := QueryHistoryResponse{
		Queries:    []*models.QueryHistory{},
		TotalCount: 0,
	}

	c.JSON(http.StatusOK, response)
}

type ValidateSQLRequest struct {
	SQLQuery string `json:"sql_query" binding:"required"`
}

type ValidateSQLResponse struct {
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
	SQLQuery         string   `json:"sql_query"`
}

type ExplainSQLRequest struct {
	SQLQuery string `json:"sql_query" binding:"required"`
}

type ExplainSQLResponse struct {
	SQLQuery    string `json:"sql_query"`
	Explanation string `json:"explanation"`
}

type SQLQueryRequest struct {
	Database     string   `json:"database" binding:"required"`
	Tables       []string `json:"tables" binding:"required"`
	Query        string   `json:"query" binding:"required"`
	ConnectionID string   `json:"connection_id,omitempty"`
	MaxRetries   int      `json:"max_retries,omitempty"`
	SaveHistory  bool     `json:"save_history,omitempty"`
}

type QueryHistoryResponse struct {
	Queries    []*models.QueryHistory        `json:"queries"`
	TotalCount int                           `json:"total_count"`
	Stats      *repository.QueryHistoryStats `json:"stats,omitempty"`
}
