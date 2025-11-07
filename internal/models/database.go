package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
	MongoDB    DatabaseType = "mongodb"
	SQLServer  DatabaseType = "sqlserver"
)

type DatabaseConnection struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name       string             `bson:"name" json:"name"`
	Type       DatabaseType       `bson:"type" json:"type"`
	Host       string             `bson:"host" json:"host"`
	Port       int                `bson:"port" json:"port"`
	Database   string             `bson:"database" json:"database"`
	Username   string             `bson:"username" json:"username"`
	Password   string             `bson:"password,omitempty" json:"-"` 
	Status     ConnectionStatus   `bson:"status" json:"status"`
	LastTested time.Time          `bson:"last_tested" json:"last_tested"`
	Version    string             `bson:"version,omitempty" json:"version,omitempty"`
	Schemas    []string           `bson:"schemas,omitempty" json:"schemas,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusError        ConnectionStatus = "error"
	StatusTesting      ConnectionStatus = "testing"
)

type DatabaseConnectionRequest struct {
	Name     string       `json:"name" binding:"required,min=1,max=50"`
	Type     DatabaseType `json:"type" binding:"required,oneof=mysql postgresql mongodb sqlserver"`
	Host     string       `json:"host" binding:"required,min=1,max=255"`
	Port     int          `json:"port" binding:"required,min=1,max=65535"`
	Database string       `json:"database" binding:"required,min=1,max=100"`
	Username string       `json:"username" binding:"required,min=1,max=100"`
	Password string       `json:"password" binding:"required,min=1"`
	Save     bool         `json:"save"`
}

type DatabaseConnectionResponse struct {
	ID       string           `json:"id,omitempty"`
	Status   ConnectionStatus `json:"status"`
	Message  string           `json:"message,omitempty"`
	Version  string           `json:"version,omitempty"`
	Schemas  []string         `json:"schemas,omitempty"`
	Database string           `json:"database,omitempty"`
}

type DatabaseSchema struct {
	Name   string        `json:"name"`
	Tables []TableSchema `json:"tables"`
}

type TableSchema struct {
	Name    string         `json:"name"`
	Columns []ColumnSchema `json:"columns"`
	Indexes []IndexSchema  `json:"indexes,omitempty"`
}

type ColumnSchema struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value,omitempty"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsAutoIncr   bool   `json:"is_auto_increment"`
	MaxLength    int    `json:"max_length,omitempty"`
}

type IndexSchema struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Primary bool     `json:"primary"`
}

type QueryRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
	Query        string `json:"query" binding:"required,min=1"`
	Limit        int    `json:"limit,omitempty"`
}

type QueryResponse struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Count   int             `json:"count"`
	Message string          `json:"message,omitempty"`
}
