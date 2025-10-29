package raft

import (
	"fmt"
	"strings"
)

func GenerateSQL(schema *Schema, provider string) (string, error) {
	var sql strings.Builder
	
	for _, model := range schema.Models {
		tableName := toSnakeCase(model.Name)
		sql.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableName))
		
		var columns []string
		var constraints []string
		
		for _, field := range model.Fields {
			colDef := generateColumn(field, provider)
			columns = append(columns, colDef)
			
			for _, attr := range field.Attributes {
				if constraint := generateConstraint(attr, field, tableName, provider); constraint != "" {
					constraints = append(constraints, constraint)
				}
			}
		}
		
		allDefs := append(columns, constraints...)
		sql.WriteString("  " + strings.Join(allDefs, ",\n  "))
		sql.WriteString("\n);\n\n")
	}
	
	return sql.String(), nil
}

func generateColumn(field Field, provider string) string {
	colName := toSnakeCase(field.Name)
	sqlType := mapTypeToSQL(field.Type, provider)
	
	var parts []string
	parts = append(parts, colName, sqlType)
	
	hasDefault := false
	isPrimary := false
	isUnique := false
	notNull := true
	
	for _, attr := range field.Attributes {
		switch attr.Name {
		case "id":
			isPrimary = true
			if provider == "postgresql" || provider == "postgres" {
				sqlType = "SERIAL"
			} else if provider == "mysql" {
				sqlType = "INT AUTO_INCREMENT"
			} else {
				sqlType = "INTEGER"
			}
		case "default":
			hasDefault = true
		case "unique":
			isUnique = true
		case "optional":
			notNull = false
		}
	}
	
	parts = []string{colName, sqlType}
	
	if isPrimary {
		parts = append(parts, "PRIMARY KEY")
	}
	
	if notNull && !isPrimary {
		parts = append(parts, "NOT NULL")
	}
	
	if isUnique && !isPrimary {
		parts = append(parts, "UNIQUE")
	}
	
	if hasDefault {
		defaultVal := getDefaultValue(field, provider)
		if defaultVal != "" {
			parts = append(parts, "DEFAULT", defaultVal)
		}
	}
	
	return strings.Join(parts, " ")
}

func generateConstraint(attr Attribute, field Field, tableName, provider string) string {
	switch attr.Name {
	case "relation":
		if attr.Value != "" {
			parts := strings.Split(attr.Value, ".")
			if len(parts) == 2 {
				refTable := toSnakeCase(parts[0])
				refCol := toSnakeCase(parts[1])
				colName := toSnakeCase(field.Name)
				return fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", colName, refTable, refCol)
			}
		}
	}
	return ""
}

func mapTypeToSQL(raftType, provider string) string {
	switch strings.ToLower(raftType) {
	case "string":
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "int", "integer":
		return "INTEGER"
	case "bigint":
		return "BIGINT"
	case "float":
		return "FLOAT"
	case "decimal":
		return "DECIMAL(10,2)"
	case "boolean", "bool":
		if provider == "postgresql" || provider == "postgres" {
			return "BOOLEAN"
		}
		return "TINYINT(1)"
	case "datetime", "timestamp":
		if provider == "postgresql" || provider == "postgres" {
			return "TIMESTAMP"
		}
		return "DATETIME"
	case "date":
		return "DATE"
	case "json":
		if provider == "mysql" {
			return "JSON"
		}
		return "TEXT"
	default:
		return "VARCHAR(255)"
	}
}

func getDefaultValue(field Field, provider string) string {
	for _, attr := range field.Attributes {
		if attr.Name == "default" {
			if attr.Value != "" {
				return attr.Value
			}
			
			switch strings.ToLower(field.Type) {
			case "string", "text":
				return "''"
			case "int", "integer", "bigint":
				return "0"
			case "boolean", "bool":
				return "false"
			case "datetime", "timestamp":
				if provider == "postgresql" || provider == "postgres" {
					return "CURRENT_TIMESTAMP"
				}
				return "CURRENT_TIMESTAMP"
			}
		}
	}
	return ""
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
