package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"gosqlpp/internal/database"

	"github.com/rodaine/table"
	"gopkg.in/yaml.v3"
)

// Formatter handles different output formats
type Formatter struct {
	format string
	writer io.Writer
}

// NewFormatter creates a new output formatter
func NewFormatter(format string, writer io.Writer) *Formatter {
	return &Formatter{
		format: format,
		writer: writer,
	}
}

// FormatResult formats and outputs the execution result
func (f *Formatter) FormatResult(result *database.ExecutionResult) error {
	if result.Error != nil {
		// Always output errors as plain text
		_, err := fmt.Fprintf(f.writer, "%s\n", database.FormatError(result))
		return err
	}

	// If no rows returned, just show the affected rows message
	if len(result.Rows) == 0 {
		message := database.FormatRowsAffected(result)
		if message != "" {
			_, err := fmt.Fprintf(f.writer, "%s\n", message)
			return err
		}
		return nil
	}

	// Format the result data based on the requested format
	switch f.format {
	case "table":
		return f.formatTable(result)
	case "json":
		return f.formatJSON(result)
	case "yaml":
		return f.formatYAML(result)
	case "csv":
		return f.formatCSV(result)
	default:
		return fmt.Errorf("unsupported output format: %s", f.format)
	}
}

// formatTable formats the result as a table
func (f *Formatter) formatTable(result *database.ExecutionResult) error {
	if len(result.Rows) == 0 {
		return nil
	}

	// Convert column names to interface{} slice
	headers := make([]interface{}, len(result.Columns))
	for i, col := range result.Columns {
		headers[i] = col
	}

	// Create table with headers
	tbl := table.New(headers...)
	tbl.WithWriter(f.writer)

	// Add rows
	for _, row := range result.Rows {
		// Convert all values to strings for table display
		stringRow := make([]interface{}, len(row))
		for i, val := range row {
			stringRow[i] = formatValue(val)
		}
		tbl.AddRow(stringRow...)
	}

	tbl.Print()

	// Add row count
	fmt.Fprintf(f.writer, "\n%s\n", database.FormatRowsAffected(result))

	return nil
}

// formatJSON formats the result as JSON
func (f *Formatter) formatJSON(result *database.ExecutionResult) error {
	// Convert rows to array of objects
	var records []map[string]interface{}

	for _, row := range result.Rows {
		record := make(map[string]interface{})
		for i, col := range result.Columns {
			if i < len(row) {
				record[col] = row[i]
			}
		}
		records = append(records, record)
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}

// formatYAML formats the result as YAML
func (f *Formatter) formatYAML(result *database.ExecutionResult) error {
	// Convert rows to array of objects
	var records []map[string]interface{}

	for _, row := range result.Rows {
		record := make(map[string]interface{})
		for i, col := range result.Columns {
			if i < len(row) {
				record[col] = row[i]
			}
		}
		records = append(records, record)
	}

	encoder := yaml.NewEncoder(f.writer)
	defer encoder.Close()
	return encoder.Encode(records)
}

// formatCSV formats the result as CSV
func (f *Formatter) formatCSV(result *database.ExecutionResult) error {
	writer := csv.NewWriter(f.writer)
	defer writer.Flush()

	// Write header
	if err := writer.Write(result.Columns); err != nil {
		return err
	}

	// Write rows
	for _, row := range result.Rows {
		stringRow := make([]string, len(row))
		for i, val := range row {
			stringRow[i] = formatValue(val)
		}
		if err := writer.Write(stringRow); err != nil {
			return err
		}
	}

	return nil
}

// formatValue converts a database value to a string representation
func formatValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetSupportedFormats returns a list of supported output formats
func GetSupportedFormats() []string {
	return []string{"table", "json", "yaml", "csv"}
}

// IsFormatSupported checks if the given format is supported
func IsFormatSupported(format string) bool {
	for _, supported := range GetSupportedFormats() {
		if strings.EqualFold(format, supported) {
			return true
		}
	}
	return false
}

// FormatData formats generic data (slice of maps) using the specified format
func (f *Formatter) FormatData(data []map[string]interface{}) error {
	if len(data) == 0 {
		_, err := fmt.Fprintf(f.writer, "No data to display\n")
		return err
	}

	switch f.format {
	case "table":
		return f.formatDataAsTable(data)
	case "json":
		return f.formatDataAsJSON(data)
	case "yaml":
		return f.formatDataAsYAML(data)
	case "csv":
		return f.formatDataAsCSV(data)
	default:
		return fmt.Errorf("unsupported output format: %s", f.format)
	}
}

// formatDataAsTable formats generic data as a table
func (f *Formatter) formatDataAsTable(data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Extract headers from the first record
	var headers []interface{}
	for key := range data[0] {
		headers = append(headers, key)
	}

	// Create table with headers
	tbl := table.New(headers...)
	tbl.WithWriter(f.writer)

	// Add rows
	for _, record := range data {
		row := make([]interface{}, len(headers))
		for i, header := range headers {
			row[i] = formatValue(record[header.(string)])
		}
		tbl.AddRow(row...)
	}

	tbl.Print()
	return nil
}

// formatDataAsJSON formats generic data as JSON
func (f *Formatter) formatDataAsJSON(data []map[string]interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatDataAsYAML formats generic data as YAML
func (f *Formatter) formatDataAsYAML(data []map[string]interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	defer encoder.Close()
	return encoder.Encode(data)
}

// formatDataAsCSV formats generic data as CSV
func (f *Formatter) formatDataAsCSV(data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	writer := csv.NewWriter(f.writer)
	defer writer.Flush()

	// Extract headers from the first record
	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}

	// Write headers
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write data rows
	for _, record := range data {
		row := make([]string, len(headers))
		for i, header := range headers {
			value := record[header]
			if value != nil {
				row[i] = fmt.Sprintf("%v", value)
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
