package raft

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Model struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name       string
	Type       string
	Attributes []Attribute
}

type Attribute struct {
	Name  string
	Value string
}

type Schema struct {
	Models []Model
}

func ParseRaftFile(filepath string) (*Schema, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open .raft file: %w", err)
	}
	defer file.Close()

	schema := &Schema{}
	scanner := bufio.NewScanner(file)
	
	var currentModel *Model
	modelRegex := regexp.MustCompile(`^model\s+(\w+)\s*=\s*\{`)
	fieldRegex := regexp.MustCompile(`^\s*(\w+)\s+(\w+)(.*)`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		
		if matches := modelRegex.FindStringSubmatch(line); matches != nil {
			if currentModel != nil {
				schema.Models = append(schema.Models, *currentModel)
			}
			currentModel = &Model{Name: matches[1]}
			continue
		}
		
		if line == "}" && currentModel != nil {
			schema.Models = append(schema.Models, *currentModel)
			currentModel = nil
			continue
		}
		
		if currentModel != nil {
			if matches := fieldRegex.FindStringSubmatch(line); matches != nil {
				field := Field{
					Name: matches[1],
					Type: matches[2],
				}
				
				attrStr := strings.TrimSpace(matches[3])
				field.Attributes = parseAttributes(attrStr)
				currentModel.Fields = append(currentModel.Fields, field)
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	
	return schema, nil
}

func parseAttributes(attrStr string) []Attribute {
	var attrs []Attribute
	attrRegex := regexp.MustCompile(`@(\w+)(?:\(([^)]*)\))?`)
	
	matches := attrRegex.FindAllStringSubmatch(attrStr, -1)
	for _, match := range matches {
		attr := Attribute{Name: match[1]}
		if len(match) > 2 {
			attr.Value = match[2]
		}
		attrs = append(attrs, attr)
	}
	
	return attrs
}
