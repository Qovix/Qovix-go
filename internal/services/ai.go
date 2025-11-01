package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Qovix/Qovix-go/internal/models"
)

type AIService struct {
	apiKey string
	client *http.Client
}

type SQLQueryRequest struct {
	Database   string                 `json:"database" binding:"required"`
	Tables     []string               `json:"tables" binding:"required"`
	Query      string                 `json:"query" binding:"required"`
	Schema     map[string]TableSchema `json:"schema,omitempty"`
	MaxRetries int                    `json:"max_retries,omitempty"`
}

type TableSchema struct {
	Columns []ColumnInfo `json:"columns"`
	Indexes []string     `json:"indexes,omitempty"`
}

type ColumnInfo struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	IsNullable   bool   `json:"is_nullable"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	DefaultValue string `json:"default_value,omitempty"`
}

type SQLQueryResponse struct {
	SQLQuery         string                 `json:"sql_query"`
	Explanation      string                 `json:"explanation"`
	Confidence       float64                `json:"confidence"`
	IsValid          bool                   `json:"is_valid"`
	ValidationErrors []string               `json:"validation_errors,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

func NewAIService(apiKey string) *AIService {
	return &AIService{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *AIService) GenerateSQLQuery(ctx context.Context, req SQLQueryRequest) (*SQLQueryResponse, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("gemini API key not configured")
	}

	prompt := s.buildPrompt(req)

	geminiResp, err := s.callGeminiAPI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}

	sqlResp, err := s.parseGeminiResponse(geminiResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	sqlResp.IsValid, sqlResp.ValidationErrors = s.ValidateSQL(sqlResp.SQLQuery)

	return sqlResp, nil
}

func (s *AIService) buildPrompt(req SQLQueryRequest) string {
	var prompt strings.Builder

	prompt.WriteString("You are an expert SQL query generator. Generate a SQL Server query based on the following information:\n\n")

	prompt.WriteString(fmt.Sprintf("Database: %s\n", req.Database))
	prompt.WriteString(fmt.Sprintf("Tables: %s\n", strings.Join(req.Tables, ", ")))
	prompt.WriteString(fmt.Sprintf("User Query: %s\n\n", req.Query))

	if len(req.Schema) > 0 {
		prompt.WriteString("Table Schema Information:\n")
		for tableName, schema := range req.Schema {
			prompt.WriteString(fmt.Sprintf("\nTable: %s\n", tableName))
			prompt.WriteString("Columns:\n")
			for _, col := range schema.Columns {
				nullable := "NOT NULL"
				if col.IsNullable {
					nullable = "NULL"
				}
				pk := ""
				if col.IsPrimaryKey {
					pk = " (PRIMARY KEY)"
				}
				prompt.WriteString(fmt.Sprintf("  - %s %s %s%s\n", col.Name, col.DataType, nullable, pk))
			}
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("Requirements:\n")
	prompt.WriteString("1. Generate ONLY a valid SQL Server query\n")
	prompt.WriteString("2. Use proper SQL Server syntax\n")
	prompt.WriteString("3. Only use SELECT, WHERE, JOIN, GROUP BY, ORDER BY, HAVING clauses\n")
	prompt.WriteString("4. Do NOT include INSERT, UPDATE, DELETE, DROP, CREATE, ALTER statements\n")
	prompt.WriteString("5. Use appropriate table and column names based on the schema\n")
	prompt.WriteString("6. Include comments in the SQL for clarity\n\n")

	prompt.WriteString("Response Format:\n")
	prompt.WriteString("SQL_QUERY: [Your SQL query here]\n")
	prompt.WriteString("EXPLANATION: [Brief explanation of what the query does]\n")
	prompt.WriteString("CONFIDENCE: [Your confidence level from 0.0 to 1.0]\n")

	return prompt.String()
}

func (s *AIService) callGeminiAPI(ctx context.Context, prompt string) (*GeminiResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", s.apiKey)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &geminiResp, nil
}

func (s *AIService) parseGeminiResponse(resp *GeminiResponse) (*SQLQueryResponse, error) {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in Gemini response")
	}

	content := resp.Candidates[0].Content.Parts[0].Text

	sqlQuery := s.extractSection(content, "SQL_QUERY:")
	explanation := s.extractSection(content, "EXPLANATION:")
	confidenceStr := s.extractSection(content, "CONFIDENCE:")

	if sqlQuery == "" {
		return nil, fmt.Errorf("no SQL query found in response")
	}

	confidence := 0.8 
	if confidenceStr != "" {
		if _, err := fmt.Sscanf(confidenceStr, "%f", &confidence); err != nil {
			confidence = 0.8 
		}
	}

	sqlQuery = s.cleanSQL(sqlQuery)

	return &SQLQueryResponse{
		SQLQuery:    sqlQuery,
		Explanation: explanation,
		Confidence:  confidence,
	}, nil
}

func (s *AIService) extractSection(content, marker string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	capturing := false

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), marker) {
			capturing = true
			if idx := strings.Index(line, marker); idx != -1 {
				remaining := strings.TrimSpace(line[idx+len(marker):])
				if remaining != "" {
					result.WriteString(remaining)
					result.WriteString("\n")
				}
			}
			continue
		}

		if capturing {
			if strings.Contains(line, ":") && (strings.HasPrefix(strings.TrimSpace(line), "SQL_QUERY:") ||
				strings.HasPrefix(strings.TrimSpace(line), "EXPLANATION:") ||
				strings.HasPrefix(strings.TrimSpace(line), "CONFIDENCE:")) {
				break
			}
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String())
}

func (s *AIService) cleanSQL(sql string) string {
	sql = strings.ReplaceAll(sql, "```sql", "")
	sql = strings.ReplaceAll(sql, "```", "")

	sql = strings.TrimSpace(sql)

	if !strings.HasSuffix(sql, ";") {
		sql += ";"
	}

	return sql
}

func (s *AIService) ValidateSQL(sql string) (bool, []string) {
	var errors []string

	cleanedSQL := s.cleanSQLForValidation(sql)
	lowerSQL := strings.ToLower(strings.TrimSpace(cleanedSQL))

	dangerousOps := []string{
		"drop", "delete", "insert", "update", "alter", "create",
		"truncate", "exec", "execute", "sp_", "xp_", "exec(",
	}

	for _, op := range dangerousOps {
		if strings.Contains(lowerSQL, op) {
			errors = append(errors, fmt.Sprintf("Dangerous operation detected: %s", op))
		}
	}

	if !strings.HasPrefix(lowerSQL, "select") {
		errors = append(errors, "Query must start with SELECT")
	}

	injectionPatterns := []string{
		"union select", "or 1=1", "or '1'='1'", "'; drop", "'; delete", "'; insert",
		"/*! ", "*/;", "' or ", "\" or ", "' and ", "\" and ",
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(lowerSQL, pattern) {
			errors = append(errors, fmt.Sprintf("Potential SQL injection pattern detected: %s", pattern))
		}
	}

	if !s.isValidSQLStructure(cleanedSQL) {
		errors = append(errors, "Invalid SQL structure")
	}

	return len(errors) == 0, errors
}

func (s *AIService) cleanSQLForValidation(sql string) string {
	lines := strings.Split(sql, "\n")
	var cleanedLines []string

	for _, line := range lines {
		if commentIndex := strings.Index(line, "--"); commentIndex != -1 {
			line = strings.TrimSpace(line[:commentIndex])
		}

		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	cleaned := strings.Join(cleanedLines, " ")

	for strings.Contains(cleaned, "/*") && strings.Contains(cleaned, "*/") {
		start := strings.Index(cleaned, "/*")
		end := strings.Index(cleaned[start:], "*/")
		if end != -1 {
			cleaned = cleaned[:start] + " " + cleaned[start+end+2:]
		} else {
			break
		}
	}

	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}

	return strings.TrimSpace(cleaned)
}

func (s *AIService) isValidSQLStructure(sql string) bool {
	cleanedSQL := strings.ToLower(strings.TrimSpace(sql))

	patterns := []string{
		`^select\s+.+\s+from\s+\w+.*$`,               
		`^select\s+[\w\s,*()]+\s+from\s+[\w\s,]+.*$`, 
		`^select\s+.*\s+from\s+.*$`,                 
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, cleanedSQL)
		if err == nil && matched {
			return true
		}
	}

	return strings.Contains(cleanedSQL, "select") && strings.Contains(cleanedSQL, "from")
}

func (s *AIService) ExplainSQL(ctx context.Context, sqlQuery string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("gemini API key not configured")
	}

	prompt := fmt.Sprintf(`Please explain the following SQL query in simple, clear terms:

SQL Query:
%s

Provide an explanation that covers:
1. What data the query retrieves
2. Which tables are involved
3. Any filtering conditions (WHERE clauses)
4. Any grouping or aggregation
5. The ordering of results

Keep the explanation concise and understandable for non-technical users.`, sqlQuery)

	geminiResp, err := s.callGeminiAPI(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in Gemini response")
	}

	explanation := geminiResp.Candidates[0].Content.Parts[0].Text
	return strings.TrimSpace(explanation), nil
}

func (s *AIService) GetTableSchema(ctx context.Context, databaseService *DatabaseService, connectionID string, tables []string) (map[string]TableSchema, error) {
	schema := make(map[string]TableSchema)

	dbSchema, err := databaseService.GetSchema(ctx, connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database schema: %w", err)
	}

	for _, tableName := range tables {
		var foundTable *models.TableSchema
		for _, table := range dbSchema.Tables {
			if strings.EqualFold(table.Name, tableName) {
				foundTable = &table
				break
			}
		}

		if foundTable == nil {
			continue 
		}

		columns := make([]ColumnInfo, len(foundTable.Columns))
		for i, col := range foundTable.Columns {
			columns[i] = ColumnInfo{
				Name:         col.Name,
				DataType:     col.Type,
				IsNullable:   col.Nullable,
				IsPrimaryKey: col.IsPrimaryKey,
				DefaultValue: col.DefaultValue,
			}
		}

		var indexes []string
		for _, index := range foundTable.Indexes {
			indexes = append(indexes, index.Name)
		}

		schema[tableName] = TableSchema{
			Columns: columns,
			Indexes: indexes,
		}
	}

	return schema, nil
}

func (s *AIService) EnhanceRequestWithSchema(ctx context.Context, req *SQLQueryRequest, databaseService *DatabaseService, connectionID string) error {
	if len(req.Tables) == 0 {
		return nil 
	}

	schema, err := s.GetTableSchema(ctx, databaseService, connectionID, req.Tables)
	if err != nil {
		fmt.Printf("Warning: Could not fetch schema information: %v\n", err)
		return nil
	}

	req.Schema = schema
	return nil
}
