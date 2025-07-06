package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gosqlpp/internal/database"
	"gosqlpp/internal/output"
	"gosqlpp/internal/preprocessor"
	"gosqlpp/internal/schema"
)

// Statement represents a parsed SQL statement with location information
type Statement struct {
	SQL       string
	StartLine int
	EndLine   int
	FileName  string
	Location  preprocessor.SourceLocation
}

// Processor handles file processing and SQL execution
type Processor struct {
	executor     *database.Executor
	formatter    *output.Formatter
	introspector *schema.Introspector
	endOnError   bool
}

// NewProcessor creates a new file processor
func NewProcessor(executor *database.Executor, formatter *output.Formatter, introspector *schema.Introspector, endOnError bool) *Processor {
	return &Processor{
		executor:     executor,
		formatter:    formatter,
		introspector: introspector,
		endOnError:   endOnError,
	}
}

// ProcessFile processes a single SQL file
func (p *Processor) ProcessFile(filename string) error {
	// Create preprocessor and process file
	prep := preprocessor.NewPreprocessor()
	lines, locations, err := prep.ProcessFile(filename)
	if err != nil {
		return fmt.Errorf("preprocessing failed for %s: %w", filename, err)
	}

	// Parse statements from preprocessed lines
	statements, err := p.parseStatementsFromLines(lines, locations)
	if err != nil {
		return err
	}

	// Execute statements
	return p.executeStatements(statements)
}

// ProcessStdin processes SQL commands from standard input
func (p *Processor) ProcessStdin() error {
	// Create preprocessor and process stdin
	prep := preprocessor.NewPreprocessor()
	lines, locations, err := prep.ProcessReader(os.Stdin, "<stdin>")
	if err != nil {
		return fmt.Errorf("preprocessing failed for stdin: %w", err)
	}

	// Parse statements from preprocessed lines
	statements, err := p.parseStatementsFromLines(lines, locations)
	if err != nil {
		return err
	}

	// Execute statements
	return p.executeStatements(statements)
}

// ProcessDirectory processes all .sql files in a directory
func (p *Processor) ProcessDirectory(dirPath string, newerThan time.Time) error {
	// Find all .sql files
	files, err := findSQLFiles(dirPath, newerThan)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Printf("No SQL files found in directory: %s\n", dirPath)
		return nil
	}

	fmt.Printf("Found %d SQL files to process\n", len(files))

	// Process each file
	for i, file := range files {
		fmt.Printf("\n[%d/%d] Processing: %s\n", i+1, len(files), file)

		if err := p.ProcessFile(file); err != nil {
			if p.endOnError {
				return fmt.Errorf("error processing %s: %w", file, err)
			}
			fmt.Printf("Error processing %s: %v\n", file, err)
		}
	}

	return nil
}

// parseStatementsFromLines parses SQL statements from preprocessed lines
func (p *Processor) parseStatementsFromLines(lines []string, locations []preprocessor.SourceLocation) ([]Statement, error) {
	var statements []Statement
	var currentStatement strings.Builder
	var startLine int
	var startLocation preprocessor.SourceLocation

	for i, line := range lines {
		// Check if this line is a schema command
		if schema.IsSchemaCommand(line) {
			// End current statement if exists
			if currentStatement.Len() > 0 {
				statements = append(statements, Statement{
					SQL:       strings.TrimSpace(currentStatement.String()),
					StartLine: startLine,
					EndLine:   i,
					FileName:  startLocation.FileName,
					Location:  startLocation,
				})
				currentStatement.Reset()
			}

			// Add schema command as a statement
			location := startLocation
			if i < len(locations) {
				location = locations[i]
			}
			statements = append(statements, Statement{
				SQL:       strings.TrimSpace(line),
				StartLine: i + 1,
				EndLine:   i + 1,
				FileName:  location.FileName,
				Location:  location,
			})
			continue
		}

		// Check if line starts with "go " (statement delimiter)
		if strings.HasPrefix(strings.TrimSpace(strings.ToLower(line)), "go ") ||
			strings.TrimSpace(strings.ToLower(line)) == "go" {
			// End current statement
			if currentStatement.Len() > 0 {
				statements = append(statements, Statement{
					SQL:       strings.TrimSpace(currentStatement.String()),
					StartLine: startLine,
					EndLine:   i,
					FileName:  startLocation.FileName,
					Location:  startLocation,
				})
				currentStatement.Reset()
			}
			continue
		}

		// Add line to current statement
		if currentStatement.Len() == 0 {
			startLine = i + 1
			if i < len(locations) {
				startLocation = locations[i]
			}
		}
		currentStatement.WriteString(line)
		currentStatement.WriteString("\n")
	}

	// Add final statement if exists
	if currentStatement.Len() > 0 {
		statements = append(statements, Statement{
			SQL:       strings.TrimSpace(currentStatement.String()),
			StartLine: startLine,
			EndLine:   len(lines),
			FileName:  startLocation.FileName,
			Location:  startLocation,
		})
	}

	return statements, nil
}

// executeStatements executes a list of SQL statements
func (p *Processor) executeStatements(statements []Statement) error {
	for _, stmt := range statements {
		// Skip empty statements
		if strings.TrimSpace(stmt.SQL) == "" {
			continue
		}

		// Check if this is a schema command
		if schema.IsSchemaCommand(stmt.SQL) {
			command, filter := schema.ParseSchemaCommand(stmt.SQL)
			if err := p.introspector.ProcessSchemaCommand(command, filter); err != nil {
				if p.endOnError {
					return fmt.Errorf("schema command error: %w", err)
				}
				fmt.Printf("Schema command error: %v\n", err)
			}
			continue
		}

		// Execute regular SQL statement using original file location for error reporting
		result := p.executor.Execute(stmt.SQL, stmt.Location.OriginalLine, stmt.Location.OriginalFile)

		// Format and output result
		if err := p.formatter.FormatResult(result); err != nil {
			return fmt.Errorf("error formatting result: %w", err)
		}

		// Check for errors
		if result.Error != nil && p.endOnError {
			return fmt.Errorf("execution stopped due to error")
		}
	}

	return nil
}

// findSQLFiles finds all .sql files in a directory, optionally filtering by modification time
func findSQLFiles(dirPath string, newerThan time.Time) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check file extension
		if strings.ToLower(filepath.Ext(path)) != ".sql" {
			return nil
		}

		// Check modification time if specified
		if !newerThan.IsZero() && info.ModTime().Before(newerThan) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// GetFileInfo returns basic information about a file
func GetFileInfo(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}
