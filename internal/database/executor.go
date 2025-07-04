package database

import (
	"fmt"
	"strings"
)

// ExecutionResult represents the result of executing a SQL statement
type ExecutionResult struct {
	RowsAffected int64
	Columns      []string
	Rows         [][]interface{}
	Error        error
	Statement    string
	LineNumber   int
	FileName     string
}

// Executor handles SQL statement execution
type Executor struct {
	connection *Connection
}

// NewExecutor creates a new SQL executor for the given connection
func NewExecutor(conn *Connection) *Executor {
	return &Executor{
		connection: conn,
	}
}

// Execute runs a SQL statement and returns the result
func (e *Executor) Execute(statement string, lineNumber int, fileName string) *ExecutionResult {
	result := &ExecutionResult{
		Statement:  statement,
		LineNumber: lineNumber,
		FileName:   fileName,
	}
	
	// Trim whitespace and check if statement is empty
	statement = strings.TrimSpace(statement)
	if statement == "" {
		return result
	}
	
	// Determine if this is a query or an execution statement
	if isQueryStatement(statement) {
		return e.executeQuery(statement, result)
	} else {
		return e.executeStatement(statement, result)
	}
}

// executeQuery executes a SELECT statement and returns rows
func (e *Executor) executeQuery(statement string, result *ExecutionResult) *ExecutionResult {
	rows, err := e.connection.DB.Query(statement)
	if err != nil {
		result.Error = err
		return result
	}
	defer rows.Close()
	
	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		result.Error = err
		return result
	}
	result.Columns = columns
	
	// Read all rows
	var allRows [][]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			result.Error = err
			return result
		}
		
		// Convert byte slices to strings for better display
		for i, val := range values {
			if b, ok := val.([]byte); ok {
				values[i] = string(b)
			}
		}
		
		allRows = append(allRows, values)
	}
	
	if err := rows.Err(); err != nil {
		result.Error = err
		return result
	}
	
	result.Rows = allRows
	return result
}

// executeStatement executes a non-query statement (INSERT, UPDATE, DELETE, etc.)
func (e *Executor) executeStatement(statement string, result *ExecutionResult) *ExecutionResult {
	sqlResult, err := e.connection.DB.Exec(statement)
	if err != nil {
		result.Error = err
		return result
	}
	
	// Get rows affected (if supported)
	if rowsAffected, err := sqlResult.RowsAffected(); err == nil {
		result.RowsAffected = rowsAffected
	}
	
	return result
}

// isQueryStatement determines if a SQL statement is a query (SELECT) or not
func isQueryStatement(statement string) bool {
	// Normalize the statement
	normalized := strings.TrimSpace(strings.ToUpper(statement))
	
	// Check for common query keywords
	queryKeywords := []string{
		"SELECT",
		"WITH",     // Common Table Expressions
		"SHOW",     // MySQL/PostgreSQL specific
		"DESCRIBE", // MySQL specific
		"DESC",     // MySQL specific
		"EXPLAIN",  // Query execution plans
	}
	
	for _, keyword := range queryKeywords {
		if strings.HasPrefix(normalized, keyword) {
			return true
		}
	}
	
	return false
}

// FormatError formats a database error with file and line information
func FormatError(result *ExecutionResult) string {
	if result.Error == nil {
		return ""
	}
	
	return fmt.Sprintf("%s:%d: error: %v", 
		result.FileName, result.LineNumber, result.Error)
}

// FormatRowsAffected formats the rows affected message
func FormatRowsAffected(result *ExecutionResult) string {
	if result.Error != nil {
		return ""
	}
	
	if len(result.Rows) > 0 {
		return fmt.Sprintf("(%d rows)", len(result.Rows))
	}
	
	if result.RowsAffected > 0 {
		return fmt.Sprintf("(%d rows affected)", result.RowsAffected)
	}
	
	return "(0 rows affected)"
}
