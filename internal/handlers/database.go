package handlers

import (
	"net/http"

	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/Qovix/Qovix-go/internal/services"
	"github.com/Qovix/Qovix-go/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DatabaseHandler struct {
	dbService *services.DatabaseService
	log       *logrus.Logger
}

func NewDatabaseHandler(dbService *services.DatabaseService) *DatabaseHandler {
	return &DatabaseHandler{
		dbService: dbService,
		log:       logger.GetLogger(),
	}
}

func (h *DatabaseHandler) TestConnection(c *gin.Context) {
	var req models.DatabaseConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Invalid connection request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid connection parameters",
			Details: map[string]interface{}{"validation_errors": err.Error()},
		})
		return
	}

	logData := map[string]interface{}{
		"type":     req.Type,
		"host":     req.Host,
		"port":     req.Port,
		"database": req.Database,
		"username": req.Username,
	}
	h.log.WithFields(logData).Info("Testing database connection")

	response, err := h.dbService.TestConnection(c.Request.Context(), &req)
	if err != nil {
		h.log.WithError(err).Error("Connection test failed")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "connection_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Connection test completed",
		"data":    response,
	})
}

func (h *DatabaseHandler) Connect(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	var req models.DatabaseConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Invalid connection request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid connection parameters",
			Details: map[string]interface{}{"validation_errors": err.Error()},
		})
		return
	}

	logData := map[string]interface{}{
		"user_id":  userID.(string),
		"type":     req.Type,
		"host":     req.Host,
		"port":     req.Port,
		"database": req.Database,
		"username": req.Username,
	}
	h.log.WithFields(logData).Info("Creating database connection")

	conn, err := h.dbService.CreateConnection(c.Request.Context(), userID.(string), &req)
	if err != nil {
		h.log.WithError(err).Error("Failed to create database connection")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "connection_failed",
			Message: err.Error(),
		})
		return
	}

	testResult, err := h.dbService.TestConnection(c.Request.Context(), &req)
	if err != nil || testResult.Status != models.StatusConnected {
		h.dbService.CloseConnection(conn.ID)

		message := "Connection failed"
		if testResult != nil && testResult.Message != "" {
			message = testResult.Message
		} else if err != nil {
			message = err.Error()
		}

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "connection_failed",
			Message: message,
		})
		return
	}

	response := &models.DatabaseConnectionResponse{
		ID:       conn.ID,
		Status:   models.StatusConnected,
		Message:  "Successfully connected to database",
		Version:  testResult.Version,
		Schemas:  testResult.Schemas,
		Database: req.Database,
	}

	h.log.WithField("connection_id", conn.ID).Info("Database connection created successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Database connected successfully",
		"data":    response,
	})
}

func (h *DatabaseHandler) Disconnect(c *gin.Context) {
	connectionID := c.Param("connection_id")
	if connectionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Connection ID is required",
		})
		return
	}

	conn, err := h.dbService.GetConnection(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "connection_not_found",
			Message: "Database connection not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || conn.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied to this connection",
		})
		return
	}

	if err := h.dbService.CloseConnection(connectionID); err != nil {
		h.log.WithError(err).Error("Failed to close database connection")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "disconnect_failed",
			Message: "Failed to close database connection",
		})
		return
	}

	h.log.WithField("connection_id", connectionID).Info("Database connection closed")

	c.JSON(http.StatusOK, gin.H{
		"message": "Database disconnected successfully",
	})
}

func (h *DatabaseHandler) GetSchema(c *gin.Context) {
	connectionID := c.Param("connection_id")
	if connectionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Connection ID is required",
		})
		return
	}

	conn, err := h.dbService.GetConnection(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "connection_not_found",
			Message: "Database connection not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || conn.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied to this connection",
		})
		return
	}

	schema, err := h.dbService.GetSchema(c.Request.Context(), connectionID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get database schema")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "schema_failed",
			Message: "Failed to retrieve database schema",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Schema retrieved successfully",
		"data":    schema,
	})
}

func (h *DatabaseHandler) ExecuteQuery(c *gin.Context) {
	connectionID := c.Param("connection_id")
	if connectionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Connection ID is required",
		})
		return
	}

	var req models.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Details: map[string]interface{}{"validation_errors": err.Error()},
		})
		return
	}

	conn, err := h.dbService.GetConnection(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "connection_not_found",
			Message: "Database connection not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || conn.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied to this connection",
		})
		return
	}

	limit := req.Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	h.log.WithFields(logrus.Fields{
		"connection_id": connectionID,
		"user_id":       userID.(string),
		"query_length":  len(req.Query),
		"limit":         limit,
	}).Info("Executing database query")

	result, err := h.dbService.ExecuteQuery(c.Request.Context(), connectionID, req.Query, limit)
	if err != nil {
		h.log.WithError(err).Error("Query execution failed")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Query executed successfully",
		"data":    result,
	})
}

func (h *DatabaseHandler) GetConnectionStatus(c *gin.Context) {
	connectionID := c.Param("connection_id")
	if connectionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Connection ID is required",
		})
		return
	}

	conn, err := h.dbService.GetConnection(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "connection_not_found",
			Message: "Database connection not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || conn.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied to this connection",
		})
		return
	}

	status := &models.DatabaseConnectionResponse{
		ID:       conn.ID,
		Status:   models.StatusConnected,
		Database: conn.Config.Database,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": status,
	})
}
