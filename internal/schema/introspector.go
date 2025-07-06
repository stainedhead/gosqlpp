package schema

import (
	"fmt"
	"strings"

	"gosqlpp/internal/database"
	"gosqlpp/internal/output"

	"github.com/jimsmart/schema"
	"github.com/schollz/progressbar/v3"
)

// Introspector handles database schema introspection
type Introspector struct {
	connection *database.Connection
	formatter  *output.Formatter
}

// NewIntrospector creates a new schema introspector
func NewIntrospector(conn *database.Connection, formatter *output.Formatter) *Introspector {
	return &Introspector{
		connection: conn,
		formatter:  formatter,
	}
}

// ProcessSchemaCommand processes @schema-* commands
func (i *Introspector) ProcessSchemaCommand(command, filter string) error {
	switch command {
	case "@schema-all":
		return i.processSchemaAll(filter)
	case "@schema-tables":
		return i.processSchemaTables(filter)
	case "@schema-views":
		return i.processSchemaViews(filter)
	case "@schema-procedures":
		return i.processSchemaProcedures(filter)
	case "@schema-functions":
		return i.processSchemaFunctions(filter)
	case "@drivers":
		return i.processDrivers(filter)
	default:
		return fmt.Errorf("unknown schema command: %s", command)
	}
}

// processSchemaAll processes @schema-all command
func (i *Introspector) processSchemaAll(filter string) error {
	fmt.Println("=== Database Schema Information ===")

	commands := []struct {
		name    string
		command string
	}{
		{"Tables", "@schema-tables"},
		{"Views", "@schema-views"},
		{"Procedures", "@schema-procedures"},
		{"Functions", "@schema-functions"},
	}

	for _, cmd := range commands {
		fmt.Printf("\n--- %s ---\n", cmd.name)
		if err := i.ProcessSchemaCommand(cmd.command, filter); err != nil {
			fmt.Printf("Error retrieving %s: %v\n", strings.ToLower(cmd.name), err)
		}
	}

	return nil
}

// processSchemaTables processes @schema-tables command
func (i *Introspector) processSchemaTables(filter string) error {
	tableNames, err := schema.TableNames(i.connection.DB)
	if err != nil {
		return fmt.Errorf("failed to retrieve table names: %w", err)
	}

	// Convert [][2]string to []string (table names only)
	var tables []string
	for _, table := range tableNames {
		// table[0] is schema, table[1] is table name
		// For SQLite, schema might be empty, so use table[1] or table[0] if table[1] is empty
		tableName := table[1]
		if tableName == "" {
			tableName = table[0]
		}
		tables = append(tables, tableName)
	}

	// Filter tables if filter is provided
	if filter != "" {
		tables = filterNames(tables, filter)
	}

	if len(tables) == 0 {
		fmt.Println("No tables found")
		return nil
	}

	// Create progress bar
	bar := progressbar.NewOptions(
		len(tables),
		progressbar.OptionSetDescription("Retrieving table info"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	// Get detailed table information
	var tableInfo []map[string]interface{}
	for _, tableName := range tables {
		bar.Add(1)

		info := map[string]interface{}{
			"table_name": tableName,
			"type":       "TABLE",
		}

		// Get column information
		columns, err := schema.ColumnTypes(i.connection.DB, "", tableName)
		if err == nil {
			info["column_count"] = len(columns)
			var columnNames []string
			for _, col := range columns {
				columnNames = append(columnNames, col.Name())
			}
			info["columns"] = strings.Join(columnNames, ", ")
		} else {
			info["column_count"] = "N/A"
			info["columns"] = "Error retrieving columns"
		}

		tableInfo = append(tableInfo, info)
	}

	bar.Finish()
	fmt.Println()

	// Format and display results
	return i.formatSchemaResults(tableInfo)
}

// processSchemaViews processes @schema-views command
func (i *Introspector) processSchemaViews(filter string) error {
	viewNames, err := schema.ViewNames(i.connection.DB)
	if err != nil {
		return fmt.Errorf("failed to retrieve view names: %w", err)
	}

	// Convert [][2]string to []string (view names only)
	var views []string
	for _, view := range viewNames {
		// view[0] is schema, view[1] is view name
		// For SQLite, schema might be empty, so use view[1] or view[0] if view[1] is empty
		viewName := view[1]
		if viewName == "" {
			viewName = view[0]
		}
		views = append(views, viewName)
	}

	// Filter views if filter is provided
	if filter != "" {
		views = filterNames(views, filter)
	}

	if len(views) == 0 {
		fmt.Println("No views found")
		return nil
	}

	// Create progress bar
	bar := progressbar.NewOptions(
		len(views),
		progressbar.OptionSetDescription("Retrieving view info"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	// Get detailed view information
	var viewInfo []map[string]interface{}
	for _, viewName := range views {
		bar.Add(1)

		info := map[string]interface{}{
			"view_name": viewName,
			"type":      "VIEW",
		}

		// Get column information
		columns, err := schema.ColumnTypes(i.connection.DB, "", viewName)
		if err == nil {
			info["column_count"] = len(columns)
			var columnNames []string
			for _, col := range columns {
				columnNames = append(columnNames, col.Name())
			}
			info["columns"] = strings.Join(columnNames, ", ")
		} else {
			info["column_count"] = "N/A"
			info["columns"] = "Error retrieving columns"
		}

		viewInfo = append(viewInfo, info)
	}

	bar.Finish()
	fmt.Println()

	// Format and display results
	return i.formatSchemaResults(viewInfo)
}

// processSchemaProcedures processes @schema-procedures command
func (i *Introspector) processSchemaProcedures(filter string) error {
	// Check if stored procedures are supported
	if !i.supportsStoredProcedures() {
		fmt.Printf("Stored procedures are not supported by %s driver\n", i.connection.Driver)
		return nil
	}

	procedures, err := i.getStoredProcedures()
	if err != nil {
		return fmt.Errorf("failed to retrieve stored procedures: %w", err)
	}

	// Filter procedures if filter is provided
	if filter != "" {
		procedures = filterNames(procedures, filter)
	}

	if len(procedures) == 0 {
		fmt.Println("No stored procedures found")
		return nil
	}

	// Create progress bar
	bar := progressbar.NewOptions(
		len(procedures),
		progressbar.OptionSetDescription("Retrieving procedure info"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	// Get detailed procedure information
	var procInfo []map[string]interface{}
	for _, procName := range procedures {
		bar.Add(1)

		info := map[string]interface{}{
			"procedure_name": procName,
			"type":           "PROCEDURE",
		}

		procInfo = append(procInfo, info)
	}

	bar.Finish()
	fmt.Println()

	// Format and display results
	return i.formatSchemaResults(procInfo)
}

// processSchemaFunctions processes @schema-functions command
func (i *Introspector) processSchemaFunctions(filter string) error {
	// Check if functions are supported
	if !i.supportsFunctions() {
		fmt.Printf("Functions are not supported by %s driver\n", i.connection.Driver)
		return nil
	}

	functions, err := i.getFunctions()
	if err != nil {
		return fmt.Errorf("failed to retrieve functions: %w", err)
	}

	// Filter functions if filter is provided
	if filter != "" {
		functions = filterNames(functions, filter)
	}

	if len(functions) == 0 {
		fmt.Println("No functions found")
		return nil
	}

	// Create progress bar
	bar := progressbar.NewOptions(
		len(functions),
		progressbar.OptionSetDescription("Retrieving function info"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	// Get detailed function information
	var funcInfo []map[string]interface{}
	for _, funcName := range functions {
		bar.Add(1)

		info := map[string]interface{}{
			"function_name": funcName,
			"type":          "FUNCTION",
		}

		funcInfo = append(funcInfo, info)
	}

	bar.Finish()
	fmt.Println()

	// Format and display results
	return i.formatSchemaResults(funcInfo)
}

// formatSchemaResults formats and displays schema results
func (i *Introspector) formatSchemaResults(results []map[string]interface{}) error {
	if len(results) == 0 {
		return nil
	}

	// Convert to database.ExecutionResult format
	var columns []string
	var rows [][]interface{}

	// Get column names from first result
	for key := range results[0] {
		columns = append(columns, key)
	}

	// Convert data to rows
	for _, result := range results {
		var row []interface{}
		for _, col := range columns {
			row = append(row, result[col])
		}
		rows = append(rows, row)
	}

	// Create execution result
	execResult := &database.ExecutionResult{
		Columns: columns,
		Rows:    rows,
	}

	// Format and display
	return i.formatter.FormatResult(execResult)
}

// filterNames filters a list of names based on a prefix filter
func filterNames(names []string, filter string) []string {
	var filtered []string
	filterLower := strings.ToLower(filter)

	for _, name := range names {
		if strings.HasPrefix(strings.ToLower(name), filterLower) {
			filtered = append(filtered, name)
		}
	}

	return filtered
}

// supportsStoredProcedures checks if the database supports stored procedures
func (i *Introspector) supportsStoredProcedures() bool {
	switch i.connection.Driver {
	case "postgres", "mysql", "sqlserver":
		return true
	case "sqlite3":
		return false
	default:
		return false
	}
}

// supportsFunctions checks if the database supports functions
func (i *Introspector) supportsFunctions() bool {
	switch i.connection.Driver {
	case "postgres", "mysql", "sqlserver":
		return true
	case "sqlite3":
		return false
	default:
		return false
	}
}

// getStoredProcedures retrieves stored procedures (database-specific)
func (i *Introspector) getStoredProcedures() ([]string, error) {
	var query string

	switch i.connection.Driver {
	case "postgres":
		query = `SELECT routine_name FROM information_schema.routines 
				WHERE routine_type = 'PROCEDURE' AND routine_schema = 'public'
				ORDER BY routine_name`
	case "mysql":
		query = `SELECT routine_name FROM information_schema.routines 
				WHERE routine_type = 'PROCEDURE' AND routine_schema = DATABASE()
				ORDER BY routine_name`
	case "sqlserver":
		query = `SELECT name FROM sys.procedures ORDER BY name`
	default:
		return nil, fmt.Errorf("stored procedures not supported for driver: %s", i.connection.Driver)
	}

	return i.executeStringQuery(query)
}

// getFunctions retrieves functions (database-specific)
func (i *Introspector) getFunctions() ([]string, error) {
	var query string

	switch i.connection.Driver {
	case "postgres":
		query = `SELECT routine_name FROM information_schema.routines 
				WHERE routine_type = 'FUNCTION' AND routine_schema = 'public'
				ORDER BY routine_name`
	case "mysql":
		query = `SELECT routine_name FROM information_schema.routines 
				WHERE routine_type = 'FUNCTION' AND routine_schema = DATABASE()
				ORDER BY routine_name`
	case "sqlserver":
		query = `SELECT name FROM sys.objects WHERE type IN ('FN', 'IF', 'TF') ORDER BY name`
	default:
		return nil, fmt.Errorf("functions not supported for driver: %s", i.connection.Driver)
	}

	return i.executeStringQuery(query)
}

// processDrivers processes @drivers command
func (i *Introspector) processDrivers(filter string) error {
	// Get supported drivers from the database package
	drivers := database.GetSupportedDrivers()

	// Filter drivers if filter is provided
	if filter != "" {
		drivers = filterNames(drivers, filter)
	}

	if len(drivers) == 0 {
		fmt.Println("No drivers found")
		return nil
	}

	// Create progress bar
	bar := progressbar.NewOptions(
		len(drivers),
		progressbar.OptionSetDescription("Retrieving driver info"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	// Get detailed driver information
	var driverInfo []map[string]interface{}
	for _, driverName := range drivers {
		bar.Add(1)

		info := map[string]interface{}{
			"driver_name": driverName,
			"type":        "DRIVER",
			"status":      "AVAILABLE",
		}

		// Add driver-specific information
		switch driverName {
		case "postgres":
			info["description"] = "PostgreSQL driver (github.com/lib/pq)"
			info["connection_example"] = "postgres://user:password@host:port/dbname?sslmode=disable"
		case "mysql":
			info["description"] = "MySQL/MariaDB driver (github.com/go-sql-driver/mysql)"
			info["connection_example"] = "username:password@tcp(127.0.0.1:3306)/database_name"
		case "sqlite3":
			info["description"] = "SQLite driver (github.com/mattn/go-sqlite3)"
			info["connection_example"] = "path/to/database.db or :memory:"
		case "sqlserver":
			info["description"] = "SQL Server driver (github.com/microsoft/go-mssqldb)"
			info["connection_example"] = "server=localhost;user id=sa;password=password;database=testdb"
		default:
			info["description"] = "Unknown driver"
			info["connection_example"] = "N/A"
		}

		driverInfo = append(driverInfo, info)
	}

	bar.Finish()
	fmt.Println()

	// Format and display results
	return i.formatSchemaResults(driverInfo)
}

// executeStringQuery executes a query and returns a slice of strings
func (i *Introspector) executeStringQuery(query string) ([]string, error) {
	rows, err := i.connection.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		results = append(results, name)
	}

	return results, rows.Err()
}

// IsSchemaCommand checks if a line contains a schema command
func IsSchemaCommand(line string) bool {
	trimmed := strings.TrimSpace(line)
	schemaCommands := []string{
		"@schema-all",
		"@schema-tables",
		"@schema-views",
		"@schema-procedures",
		"@schema-functions",
		"@drivers",
	}

	for _, cmd := range schemaCommands {
		if strings.HasPrefix(trimmed, cmd) {
			return true
		}
	}

	return false
}

// ParseSchemaCommand parses a schema command line and returns command and filter
func ParseSchemaCommand(line string) (string, string) {
	trimmed := strings.TrimSpace(line)
	parts := strings.Fields(trimmed)

	if len(parts) == 0 {
		return "", ""
	}

	command := parts[0]
	filter := ""

	// Look for filter parameter
	if len(parts) > 1 {
		// Remove quotes if present
		filter = strings.Trim(parts[1], `"'`)
	}

	return command, filter
}
