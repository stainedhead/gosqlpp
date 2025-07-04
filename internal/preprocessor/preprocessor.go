package preprocessor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SourceLocation tracks the original location of a line for error reporting
type SourceLocation struct {
	FileName     string
	LineNumber   int
	OriginalFile string
	OriginalLine int
}

// Define represents a preprocessor #define
type Define struct {
	Name  string
	Value string
}

// Preprocessor handles SQL preprocessing with #define, #include, and conditionals
type Preprocessor struct {
	defines          map[string]Define
	locations        []SourceLocation
	conditionalStack *ConditionalStack
}

// NewPreprocessor creates a new preprocessor instance
func NewPreprocessor() *Preprocessor {
	return &Preprocessor{
		defines:   make(map[string]Define),
		locations: make([]SourceLocation, 0),
	}
}

// ProcessFile processes a file with preprocessing directives
func (p *Preprocessor) ProcessFile(filename string) ([]string, []SourceLocation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()
	
	return p.processReader(file, filename)
}

// ProcessReader processes content from an io.Reader (for stdin support)
func (p *Preprocessor) ProcessReader(reader io.Reader, filename string) ([]string, []SourceLocation, error) {
	return p.processReader(reader, filename)
}

// processReader processes content from a reader
func (p *Preprocessor) processReader(reader io.Reader, filename string) ([]string, []SourceLocation, error) {
	var result []string
	var locations []SourceLocation
	
	scanner := bufio.NewScanner(reader)
	lineNumber := 0
	
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		
		// Process the line with conditional support
		processedLines, lineLocations, err := p.processLineWithConditionals(line, filename, lineNumber)
		if err != nil {
			return nil, nil, err
		}
		
		result = append(result, processedLines...)
		locations = append(locations, lineLocations...)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}
	
	// Validate that all conditional blocks are closed
	if err := p.ValidateConditionals(filename); err != nil {
		return nil, nil, err
	}
	
	return result, locations, nil
}

// processLine processes a single line with preprocessing directives
func (p *Preprocessor) processLine(line, filename string, lineNumber int) ([]string, []SourceLocation, error) {
	trimmed := strings.TrimSpace(line)
	
	// Handle #define
	if strings.HasPrefix(trimmed, "#define ") {
		return p.processDefine(trimmed, filename, lineNumber)
	}
	
	// Handle #include
	if strings.HasPrefix(trimmed, "#include ") {
		return p.processInclude(trimmed, filename, lineNumber)
	}
	
	// Regular line - apply variable substitution
	processedLine := p.substituteVariables(line)
	location := SourceLocation{
		FileName:     filename,
		LineNumber:   lineNumber,
		OriginalFile: filename,
		OriginalLine: lineNumber,
	}
	
	return []string{processedLine}, []SourceLocation{location}, nil
}

// processDefine handles #define directives
func (p *Preprocessor) processDefine(line, filename string, lineNumber int) ([]string, []SourceLocation, error) {
	// Parse #define NAME VALUE [// comment]
	re := regexp.MustCompile(`^#define\s+(\w+)\s+(.+?)(?:\s*//.*)?$`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 3 {
		return nil, nil, fmt.Errorf("%s:%d: invalid #define syntax", filename, lineNumber)
	}
	
	name := matches[1]
	value := strings.TrimSpace(matches[2])
	
	// Remove quotes if present
	if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
		(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
		value = value[1 : len(value)-1]
	}
	
	p.defines[name] = Define{
		Name:  name,
		Value: value,
	}
	
	// #define lines are not included in output
	return []string{}, []SourceLocation{}, nil
}

// processInclude handles #include directives
func (p *Preprocessor) processInclude(line, filename string, lineNumber int) ([]string, []SourceLocation, error) {
	// Parse #include "filename" [// comment]
	re := regexp.MustCompile(`^#include\s+"([^"]+)"(?:\s*//.*)?$`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 2 {
		return nil, nil, fmt.Errorf("%s:%d: invalid #include syntax", filename, lineNumber)
	}
	
	includeFile := matches[1]
	
	// Resolve relative path
	if !filepath.IsAbs(includeFile) {
		baseDir := filepath.Dir(filename)
		includeFile = filepath.Join(baseDir, includeFile)
	}
	
	// Process the included file
	includedLines, includedLocations, err := p.ProcessFile(includeFile)
	if err != nil {
		return nil, nil, fmt.Errorf("%s:%d: error including file: %w", filename, lineNumber, err)
	}
	
	return includedLines, includedLocations, nil
}

// substituteVariables replaces #define variables in a line
func (p *Preprocessor) substituteVariables(line string) string {
	result := line
	
	for name, define := range p.defines {
		// Replace whole word matches only
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
		result = re.ReplaceAllString(result, define.Value)
	}
	
	return result
}

// GetDefines returns a copy of current defines
func (p *Preprocessor) GetDefines() map[string]Define {
	result := make(map[string]Define)
	for k, v := range p.defines {
		result[k] = v
	}
	return result
}

// SetDefine sets a define programmatically
func (p *Preprocessor) SetDefine(name, value string) {
	p.defines[name] = Define{
		Name:  name,
		Value: value,
	}
}

// HasDefine checks if a define exists
func (p *Preprocessor) HasDefine(name string) bool {
	_, exists := p.defines[name]
	return exists
}

// ClearDefines clears all defines
func (p *Preprocessor) ClearDefines() {
	p.defines = make(map[string]Define)
}
