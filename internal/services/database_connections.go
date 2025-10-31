package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/go-sql-driver/mysql"
)


func (s *DatabaseService) testSQLServerConnection(ctx context.Context, req *models.DatabaseConnectionRequest) (*models.DatabaseConnectionResponse, error) {
	dsn := fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s",
		req.Host, req.Port, req.Username, req.Password, req.Database)

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return &models.DatabaseConnectionResponse{
			Status:  models.StatusError,
			Message: fmt.Sprintf("Failed to open SQL Server connection: %v", err),
		}, nil
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return &models.DatabaseConnectionResponse{
			Status:  models.StatusError,
			Message: fmt.Sprintf("Failed to ping SQL Server: %v", err),
		}, nil
	}

	return &models.DatabaseConnectionResponse{
		Status:   models.StatusConnected,
		Message:  "Successfully connected to SQL Server database",
		Database: req.Database,
		Schemas:  []string{req.Database},
	}, nil
}

func (s *DatabaseService) getSQLServerSchema(ctx context.Context, conn *DatabaseConnection) (*models.DatabaseSchema, error) {
	db := conn.SQLConn
	if db == nil {
		return nil, fmt.Errorf("no active database connection")
	}

	s.log.WithField("database", conn.Config.Database).Info("Starting SQL Server schema extraction")

	query := `
		SELECT 
			TABLE_SCHEMA,
			TABLE_NAME,
			TABLE_TYPE
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_TYPE = 'BASE TABLE'
			AND TABLE_CATALOG = DB_NAME()
		ORDER BY TABLE_SCHEMA, TABLE_NAME
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database tables: %w", err)
	}
	defer rows.Close()

	var tables []models.TableSchema
	for rows.Next() {
		var schema, tableName, tableType string
		if err := rows.Scan(&schema, &tableName, &tableType); err != nil {
			continue
		}

		columns, err := s.getSQLServerTableColumns(ctx, db, schema, tableName)
		if err != nil {
			s.log.WithError(err).WithFields(map[string]interface{}{
				"schema": schema,
				"table":  tableName,
			}).Warn("Failed to get table columns")
			columns = []models.ColumnSchema{}
		}

		tables = append(tables, models.TableSchema{
			Name:    tableName,
			Columns: columns,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	s.log.WithFields(map[string]interface{}{
		"database":    conn.Config.Database,
		"table_count": len(tables),
	}).Info("Successfully extracted SQL Server schema")

	if len(tables) == 0 {
		s.log.WithField("database", conn.Config.Database).Warn("No tables found in SQL Server database")
	}

	return &models.DatabaseSchema{
		Name:   conn.Config.Database,
		Tables: tables,
	}, nil
}

func (s *DatabaseService) getSQLServerTableColumns(ctx context.Context, db *sql.DB, schema, tableName string) ([]models.ColumnSchema, error) {
	s.log.WithFields(map[string]interface{}{
		"schema":    schema,
		"tableName": tableName,
	}).Debug("Querying SQL Server table columns")

	query := fmt.Sprintf(`
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			CHARACTER_MAXIMUM_LENGTH
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_CATALOG = DB_NAME()
			AND TABLE_SCHEMA = '%s'
			AND TABLE_NAME = '%s'
		ORDER BY ORDINAL_POSITION
	`, schema, tableName)

	s.log.WithField("query", query).Debug("Executing column query")
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table columns: %w", err)
	}
	defer rows.Close()

	var columns []models.ColumnSchema
	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault, maxLength sql.NullString

		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &maxLength); err != nil {
			s.log.WithError(err).Error("Failed to scan column row")
			continue
		}

		column := models.ColumnSchema{
			Name:     columnName,
			Type:     dataType,
			Nullable: isNullable == "YES",
		}

		if columnDefault.Valid && columnDefault.String != "NULL" {
			column.DefaultValue = columnDefault.String
		}

		if maxLength.Valid {
			if len, err := strconv.Atoi(maxLength.String); err == nil {
				column.MaxLength = len
			}
		}

		columns = append(columns, column)
	}

	s.log.WithFields(map[string]interface{}{
		"schema":       schema,
		"tableName":    tableName,
		"column_count": len(columns),
	}).Debug("Retrieved table columns")

	return columns, rows.Err()
}

func (s *DatabaseService) createSQLConnection(req *models.DatabaseConnectionRequest) (*sql.DB, error) {
	var dsn string
	var driver string

	switch req.Type {
	case models.MySQL:
		config := mysql.NewConfig()
		config.User = req.Username
		config.Passwd = req.Password
		config.Net = "tcp"
		config.Addr = fmt.Sprintf("%s:%d", req.Host, req.Port)
		config.DBName = req.Database
		config.ParseTime = true
		dsn = config.FormatDSN()
		driver = "mysql"
	case models.PostgreSQL:
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			req.Host, req.Port, req.Username, req.Password, req.Database)
		driver = "postgres"
	case models.SQLServer:
		dsn = fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s",
			req.Host, req.Port, req.Username, req.Password, req.Database)
		driver = "sqlserver"
	default:
		return nil, fmt.Errorf("unsupported SQL database type: %s", req.Type)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetMaxOpenConns(maxOpenConnections)
	db.SetConnMaxLifetime(connMaxLifetime)

	return db, nil
}

func (s *DatabaseService) executeSQLQuery(ctx context.Context, conn *DatabaseConnection, query string, limit int) (*models.QueryResponse, error) {
	db := conn.SQLConn

	query = strings.TrimSpace(query)
	lowerQuery := strings.ToLower(query)
	if !strings.Contains(lowerQuery, "limit") &&
		!strings.Contains(lowerQuery, "top") &&
		!strings.Contains(lowerQuery, "fetch") {

		switch conn.Type {
		case models.SQLServer:
			break
		case models.MySQL, models.PostgreSQL:
			query = fmt.Sprintf("%s LIMIT %d", query, limit)
		default:
			query = fmt.Sprintf("%s LIMIT %d", query, limit)
		}
	}

	s.log.WithFields(map[string]interface{}{
		"database_type": conn.Type,
		"query":         query,
	}).Debug("Executing SQL query")

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		s.log.WithError(err).WithField("query", query).Error("Query execution failed")
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var result [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
		}

		result = append(result, values)
	}

	return &models.QueryResponse{
		Columns: columns,
		Rows:    result,
		Count:   len(result),
	}, nil
}
