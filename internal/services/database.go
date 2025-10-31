package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/Qovix/Qovix-go/pkg/logger"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
)

type DatabaseService struct {
	connections map[string]*DatabaseConnection
	mutex       sync.RWMutex
	encryptKey  []byte
	log         *logrus.Logger
}

type DatabaseConnection struct {
	ID        string
	UserID    string
	Config    *models.DatabaseConnectionRequest
	SQLConn   *sql.DB
	MongoConn *mongo.Client
	LastUsed  time.Time
	Type      models.DatabaseType
}

const (
	maxConnections     = 50
	connectionTimeout  = 30 * time.Minute
	maxIdleConnections = 5
	maxOpenConnections = 20
	connMaxLifetime    = 1 * time.Hour
)

func NewDatabaseService(encryptionKey string) *DatabaseService {
	key := sha256.Sum256([]byte(encryptionKey))

	service := &DatabaseService{
		connections: make(map[string]*DatabaseConnection),
		encryptKey:  key[:],
		log:         logger.GetLogger(),
	}

	go service.cleanupConnections()

	return service
}

func (s *DatabaseService) TestConnection(ctx context.Context, req *models.DatabaseConnectionRequest) (*models.DatabaseConnectionResponse, error) {
	s.log.WithFields(logrus.Fields{
		"type": req.Type,
		"host": req.Host,
		"port": req.Port,
		"db":   req.Database,
		"user": req.Username,
	}).Info("Testing database connection")

	switch req.Type {
	case models.SQLServer:
		return s.testSQLServerConnection(ctx, req)
	case models.MySQL:
		return &models.DatabaseConnectionResponse{
			Status:  models.StatusError,
			Message: "MySQL support is coming soon. Only SQL Server is currently available.",
		}, nil
	case models.PostgreSQL:
		return &models.DatabaseConnectionResponse{
			Status:  models.StatusError,
			Message: "PostgreSQL support is coming soon. Only SQL Server is currently available.",
		}, nil
	case models.MongoDB:
		return &models.DatabaseConnectionResponse{
			Status:  models.StatusError,
			Message: "MongoDB support is coming soon. Only SQL Server is currently available.",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", req.Type)
	}
}

func (s *DatabaseService) CreateConnection(ctx context.Context, userID string, req *models.DatabaseConnectionRequest) (*DatabaseConnection, error) {
	testResult, err := s.TestConnection(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	if testResult.Status != models.StatusConnected {
		return nil, fmt.Errorf("connection failed: %s", testResult.Message)
	}

	connectionID := s.generateConnectionID(userID, req)

	conn := &DatabaseConnection{
		ID:       connectionID,
		UserID:   userID,
		Config:   req,
		LastUsed: time.Now(),
		Type:     req.Type,
	}

	switch req.Type {
	case models.SQLServer:
		sqlConn, err := s.createSQLConnection(req)
		if err != nil {
			return nil, fmt.Errorf("failed to create SQL connection: %w", err)
		}
		conn.SQLConn = sqlConn
	case models.MySQL, models.PostgreSQL:
		return nil, fmt.Errorf("only SQL Server connections are currently supported")
	case models.MongoDB:
		return nil, fmt.Errorf("only SQL Server connections are currently supported")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", req.Type)
	}

	s.mutex.Lock()
	s.connections[connectionID] = conn
	s.mutex.Unlock()

	s.log.WithField("connection_id", connectionID).Info("Database connection created successfully")

	return conn, nil
}

func (s *DatabaseService) GetConnection(connectionID string) (*DatabaseConnection, error) {
	s.mutex.RLock()
	conn, exists := s.connections[connectionID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	conn.LastUsed = time.Now()

	return conn, nil
}

func (s *DatabaseService) CloseConnection(connectionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	conn, exists := s.connections[connectionID]
	if !exists {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	if conn.SQLConn != nil {
		conn.SQLConn.Close()
	}
	if conn.MongoConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		conn.MongoConn.Disconnect(ctx)
	}

	delete(s.connections, connectionID)
	s.log.WithField("connection_id", connectionID).Info("Database connection closed")

	return nil
}

func (s *DatabaseService) GetSchema(ctx context.Context, connectionID string) (*models.DatabaseSchema, error) {
	conn, err := s.GetConnection(connectionID)
	if err != nil {
		return nil, err
	}

	switch conn.Type {
	case models.SQLServer:
		return s.getSQLServerSchema(ctx, conn)
	case models.MySQL:
		return nil, fmt.Errorf("MySQL schema extraction is not available in this version. Only SQL Server is supported")
	case models.PostgreSQL:
		return nil, fmt.Errorf("PostgreSQL schema extraction is not available in this version. Only SQL Server is supported")
	case models.MongoDB:
		return nil, fmt.Errorf("MongoDB schema extraction is not available in this version. Only SQL Server is supported")
	default:
		return nil, fmt.Errorf("unsupported database type for schema extraction: %s", conn.Type)
	}
}

func (s *DatabaseService) ExecuteQuery(ctx context.Context, connectionID string, query string, limit int) (*models.QueryResponse, error) {
	conn, err := s.GetConnection(connectionID)
	if err != nil {
		return nil, err
	}

	if err := s.validateQuery(query); err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	if limit <= 0 {
		limit = 1000
	}

	switch conn.Type {
	case models.SQLServer:
		return s.executeSQLQuery(ctx, conn, query, limit)
	case models.MySQL, models.PostgreSQL:
		return nil, fmt.Errorf("query execution is not available for %s in this version. Only SQL Server is supported", conn.Type)
	case models.MongoDB:
		return nil, fmt.Errorf("MongoDB queries are not available in this version. Only SQL Server is supported")
	default:
		return nil, fmt.Errorf("unsupported database type for queries: %s", conn.Type)
	}
}

func (s *DatabaseService) encryptPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	plaintext := []byte(password)
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *DatabaseService) decryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("invalid encrypted data")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (s *DatabaseService) generateConnectionID(userID string, req *models.DatabaseConnectionRequest) string {
	data := fmt.Sprintf("%s:%s:%s:%d:%s", userID, req.Type, req.Host, req.Port, req.Database)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("conn_%x", hash[:8])
}

func (s *DatabaseService) cleanupConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		for id, conn := range s.connections {
			if time.Since(conn.LastUsed) > connectionTimeout {
				if conn.SQLConn != nil {
					conn.SQLConn.Close()
				}
				if conn.MongoConn != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					conn.MongoConn.Disconnect(ctx)
					cancel()
				}
				delete(s.connections, id)
				s.log.WithField("connection_id", id).Info("Cleaned up expired connection")
			}
		}
		s.mutex.Unlock()
	}
}

func (s *DatabaseService) validateQuery(query string) error {
	query = strings.TrimSpace(strings.ToLower(query))

	if !strings.HasPrefix(query, "select") {
		return fmt.Errorf("only SELECT queries are allowed")
	}

	dangerousKeywords := []string{
		"drop", "delete", "insert", "update", "alter", "create",
		"truncate", "exec", "execute", "sp_", "xp_", "into outfile",
		"load_file", "benchmark", "sleep",
	}

	for _, keyword := range dangerousKeywords {
		if strings.Contains(query, keyword) {
			return fmt.Errorf("query contains forbidden keyword: %s", keyword)
		}
	}

	return nil
}
