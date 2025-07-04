package preprocessor

import (
	"fmt"
	"regexp"
	"strings"
)

// ConditionalBlock represents a conditional preprocessing block
type ConditionalBlock struct {
	Type      string // "ifdef" or "ifndef"
	Variable  string
	StartLine int
	Active    bool
}

// ConditionalStack manages nested conditional blocks
type ConditionalStack struct {
	blocks []ConditionalBlock
}

// NewConditionalStack creates a new conditional stack
func NewConditionalStack() *ConditionalStack {
	return &ConditionalStack{
		blocks: make([]ConditionalBlock, 0),
	}
}

// Push adds a new conditional block to the stack
func (cs *ConditionalStack) Push(block ConditionalBlock) {
	cs.blocks = append(cs.blocks, block)
}

// Pop removes the top conditional block from the stack
func (cs *ConditionalStack) Pop() (ConditionalBlock, error) {
	if len(cs.blocks) == 0 {
		return ConditionalBlock{}, fmt.Errorf("no conditional blocks to pop")
	}
	
	block := cs.blocks[len(cs.blocks)-1]
	cs.blocks = cs.blocks[:len(cs.blocks)-1]
	return block, nil
}

// IsEmpty returns true if the stack is empty
func (cs *ConditionalStack) IsEmpty() bool {
	return len(cs.blocks) == 0
}

// ShouldInclude returns true if the current line should be included based on active conditionals
func (cs *ConditionalStack) ShouldInclude() bool {
	for _, block := range cs.blocks {
		if !block.Active {
			return false
		}
	}
	return true
}

// Depth returns the current nesting depth
func (cs *ConditionalStack) Depth() int {
	return len(cs.blocks)
}

// Add conditional stack to preprocessor
func (p *Preprocessor) initConditionals() {
	if p.conditionalStack == nil {
		p.conditionalStack = NewConditionalStack()
	}
}

// Update the preprocessor struct to include conditional stack
type PreprocessorWithConditionals struct {
	*Preprocessor
	conditionalStack *ConditionalStack
}

// processIfdef handles #ifdef directives
func (p *Preprocessor) processIfdef(line, filename string, lineNumber int) ([]string, []SourceLocation, error) {
	p.initConditionals()
	
	// Parse #ifdef VARIABLE_NAME [// comment]
	re := regexp.MustCompile(`^#ifdef\s+(\w+)(?:\s*//.*)?$`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 2 {
		return nil, nil, fmt.Errorf("%s:%d: invalid #ifdef syntax", filename, lineNumber)
	}
	
	variableName := matches[1]
	active := p.HasDefine(variableName)
	
	// If we're already in an inactive block, this block is also inactive
	if !p.conditionalStack.ShouldInclude() {
		active = false
	}
	
	block := ConditionalBlock{
		Type:      "ifdef",
		Variable:  variableName,
		StartLine: lineNumber,
		Active:    active,
	}
	
	p.conditionalStack.Push(block)
	
	// #ifdef lines are not included in output
	return []string{}, []SourceLocation{}, nil
}

// processIfndef handles #ifndef directives
func (p *Preprocessor) processIfndef(line, filename string, lineNumber int) ([]string, []SourceLocation, error) {
	p.initConditionals()
	
	// Parse #ifndef VARIABLE_NAME [// comment]
	re := regexp.MustCompile(`^#ifndef\s+(\w+)(?:\s*//.*)?$`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 2 {
		return nil, nil, fmt.Errorf("%s:%d: invalid #ifndef syntax", filename, lineNumber)
	}
	
	variableName := matches[1]
	active := !p.HasDefine(variableName)
	
	// If we're already in an inactive block, this block is also inactive
	if !p.conditionalStack.ShouldInclude() {
		active = false
	}
	
	block := ConditionalBlock{
		Type:      "ifndef",
		Variable:  variableName,
		StartLine: lineNumber,
		Active:    active,
	}
	
	p.conditionalStack.Push(block)
	
	// #ifndef lines are not included in output
	return []string{}, []SourceLocation{}, nil
}

// processEnd handles #end directives
func (p *Preprocessor) processEnd(line, filename string, lineNumber int) ([]string, []SourceLocation, error) {
	p.initConditionals()
	
	if p.conditionalStack.IsEmpty() {
		return nil, nil, fmt.Errorf("%s:%d: #end without matching #ifdef or #ifndef", filename, lineNumber)
	}
	
	_, err := p.conditionalStack.Pop()
	if err != nil {
		return nil, nil, fmt.Errorf("%s:%d: error processing #end: %w", filename, lineNumber, err)
	}
	
	// #end lines are not included in output
	return []string{}, []SourceLocation{}, nil
}

// Update processLine to handle conditionals
func (p *Preprocessor) processLineWithConditionals(line, filename string, lineNumber int) ([]string, []SourceLocation, error) {
	p.initConditionals()
	
	trimmed := strings.TrimSpace(line)
	
	// Handle #ifdef
	if strings.HasPrefix(trimmed, "#ifdef ") {
		return p.processIfdef(trimmed, filename, lineNumber)
	}
	
	// Handle #ifndef
	if strings.HasPrefix(trimmed, "#ifndef ") {
		return p.processIfndef(trimmed, filename, lineNumber)
	}
	
	// Handle #end
	if trimmed == "#end" {
		return p.processEnd(trimmed, filename, lineNumber)
	}
	
	// Check if we should include this line based on active conditionals
	if !p.conditionalStack.ShouldInclude() {
		// Skip this line - it's in an inactive conditional block
		return []string{}, []SourceLocation{}, nil
	}
	
	// Process normally
	return p.processLine(line, filename, lineNumber)
}

// ValidateConditionals checks if all conditional blocks are properly closed
func (p *Preprocessor) ValidateConditionals(filename string) error {
	p.initConditionals()
	
	if !p.conditionalStack.IsEmpty() {
		return fmt.Errorf("%s: unclosed conditional blocks (missing #end)", filename)
	}
	
	return nil
}
