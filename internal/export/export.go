package export

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Rana718/Graft/internal/database"
	"github.com/Rana718/Graft/internal/types"
	_ "github.com/mattn/go-sqlite3"
)

func PerformExport(ctx context.Context, adapter database.DatabaseAdapter, exportPath, format string) (string, error) {
	tables, err := adapter.GetAllTableNames(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get table names: %w", err)
	}

	if len(tables) == 0 {
		log.Println("No tables found in database")
		return "", nil
	}

	exportData := types.BackupData{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Version:   "1.0",
		Tables:    make(map[string]interface{}, len(tables)),
		Comment:   "Database export",
	}

	for _, tableName := range tables {
		if tableName == "_graft_migrations" {
			continue
		}

		if tableData, err := adapter.GetTableData(ctx, tableName); err != nil {
			log.Printf("Warning: Failed to get data for table %s: %v", tableName, err)
		} else {
			exportData.Tables[tableName] = tableData
		}
	}

	switch format {
	case "csv":
		return exportToCSV(exportData, exportPath)
	case "sqlite":
		return exportToSQLite(ctx, adapter, exportData, exportPath)
	default:
		return exportToJSON(exportData, exportPath)
	}
}

func exportToJSON(data types.BackupData, exportPath string) (string, error) {
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filePath := filepath.Join(exportPath, fmt.Sprintf("export_%s.json", timestamp))

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

func exportToCSV(data types.BackupData, exportPath string) (string, error) {
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	dirPath := filepath.Join(exportPath, fmt.Sprintf("export_%s_csv", timestamp))
	
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create CSV directory: %w", err)
	}

	for tableName, tableData := range data.Tables {
		rows, ok := tableData.([]map[string]interface{})
		if !ok || len(rows) == 0 {
			continue
		}

		filePath := filepath.Join(dirPath, fmt.Sprintf("%s.csv", tableName))
		file, err := os.Create(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to create CSV file for %s: %w", tableName, err)
		}

		writer := csv.NewWriter(file)
		
		var headers []string
		for key := range rows[0] {
			headers = append(headers, key)
		}
		writer.Write(headers)

		for _, row := range rows {
			var values []string
			for _, header := range headers {
				values = append(values, fmt.Sprintf("%v", row[header]))
			}
			writer.Write(values)
		}

		writer.Flush()
		file.Close()
	}

	return dirPath, nil
}

func exportToSQLite(ctx context.Context, adapter database.DatabaseAdapter, data types.BackupData, exportPath string) (string, error) {
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filePath := filepath.Join(exportPath, fmt.Sprintf("export_%s.db", timestamp))

	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create SQLite database: %w", err)
	}
	defer db.Close()

	for tableName, tableData := range data.Tables {
		rows, ok := tableData.([]map[string]interface{})
		if !ok || len(rows) == 0 {
			continue
		}

		var columns []string
		for key := range rows[0] {
			columns = append(columns, key)
		}

		createSQL := fmt.Sprintf("CREATE TABLE %s (%s)", tableName, buildColumnDefs(columns))
		if _, err := db.Exec(createSQL); err != nil {
			return "", fmt.Errorf("failed to create table %s: %w", tableName, err)
		}

		for _, row := range rows {
			insertSQL := buildInsertSQL(tableName, columns)
			values := make([]interface{}, len(columns))
			for i, col := range columns {
				values[i] = row[col]
			}
			if _, err := db.Exec(insertSQL, values...); err != nil {
				log.Printf("Warning: Failed to insert row into %s: %v", tableName, err)
			}
		}
	}

	return filePath, nil
}

func buildColumnDefs(columns []string) string {
	var defs []string
	for _, col := range columns {
		defs = append(defs, fmt.Sprintf("%s TEXT", col))
	}
	result := ""
	for i, def := range defs {
		if i > 0 {
			result += ", "
		}
		result += def
	}
	return result
}

func buildInsertSQL(table string, columns []string) string {
	placeholders := ""
	colNames := ""
	for i, col := range columns {
		if i > 0 {
			placeholders += ", "
			colNames += ", "
		}
		placeholders += "?"
		colNames += col
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, colNames, placeholders)
}
