package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// ANSI color codes
const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Reset  = "\033[0m"
)

// (?is)create.*?\(.*?\);

func extractTableName(createStatement string) string {
	re := regexp.MustCompile(`(?i)create\s+table\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	matches := re.FindStringSubmatch(createStatement)
	if len(matches) > 1 {
		return strings.ToLower(matches[1])
	}
	return ""
}

func sortByTableName(statements []string) {
	sort.Slice(statements, func(i, j int) bool {
		tableI := extractTableName(statements[i])
		tableJ := extractTableName(statements[j])
		return tableI < tableJ
	})
}

func findLCS(a, b []string) [][]int {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if strings.TrimSpace(a[i-1]) == strings.TrimSpace(b[j-1]) {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}
	return dp
}

func printDiff(schema1, schema2 []string, filename1, filename2 string) {
	fmt.Printf("%s=== SCHEMA COMPARISON ===%s\n", Blue, Reset)
	fmt.Printf("--- %s%s%s\n", Yellow, filename1, Reset)
	fmt.Printf("+++ %s%s%s\n", Yellow, filename2, Reset)
	fmt.Println()

	if len(schema1) == 0 && len(schema2) == 0 {
		fmt.Printf("%s Both schemas are empty%s\n", Green, Reset)
		return
	}

	if len(schema1) == len(schema2) {
		identical := true
		for i := range schema1 {
			if strings.TrimSpace(schema1[i]) != strings.TrimSpace(schema2[i]) {
				identical = false
				break
			}
		}
		if identical {
			fmt.Printf("%s Schemas are identical%s\n", Green, Reset)
			return
		}
	}

	dp := findLCS(schema1, schema2)
	i, j := len(schema1), len(schema2)

	var diff []string

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && strings.TrimSpace(schema1[i-1]) == strings.TrimSpace(schema2[j-1]) {
			diff = append([]string{" " + schema1[i-1]}, diff...)
			i--
			j--
		} else if i > 0 && (j == 0 || dp[i-1][j] >= dp[i][j-1]) {
			diff = append([]string{"-" + schema1[i-1]}, diff...)
			i--
		} else {
			diff = append([]string{"+" + schema2[j-1]}, diff...)
			j--
		}
	}

	// Print diff with colors
	for _, line := range diff {
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '-':
			fmt.Printf("%s-%s%s\n", Red, line[1:], Reset)
		case '+':
			fmt.Printf("%s+%s%s\n", Green, line[1:], Reset)
		default:
			fmt.Printf(" %s\n", line[1:])
		}
	}
}

func splitCreate(s string) []string {
	re := regexp.MustCompile(`(?is)create.*?\(.*?\);`)
	return re.FindAllString(s, -1)
}

func writeFile(filename string, data []string) error {
	return os.WriteFile(filename, []byte(strings.Join(data, "\n")), 0644)
}

func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	value := string(data)
	re := regexp.MustCompile(` {2,}`)
	result := re.ReplaceAllString(value, " ")
	re = regexp.MustCompile(`\n{2,} {2,}`)
	result = re.ReplaceAllString(result, "\n")
	return result, nil
}

func createTemp() {
	os.MkdirAll("temp", 0755)
}

func deleteTemp() {
	os.RemoveAll("temp")
}

func main() {
	createTemp()
	// defer deleteTemp()
	args := os.Args
	if len(args) < 3 {
		fmt.Printf("%sUsage: go run main.go <schema1.sql> <schema2.sql>%s\n", Red, Reset)
		return
	}

	schema_file1, err := readFile(args[1])
	if err != nil {
		fmt.Printf("%sError reading %s: %v%s\n", Red, args[1], err, Reset)
		return
	}

	schema_file2, err := readFile(args[2])
	if err != nil {
		fmt.Printf("%sError reading %s: %v%s\n", Red, args[2], err, Reset)
		return
	}

	schema1 := splitCreate(schema_file1)
	schema2 := splitCreate(schema_file2)

	// Sort both schemas by table name for consistent comparison
	sortByTableName(schema1)
	sortByTableName(schema2)

	writeFile("temp/schema1.sql", schema1)
	writeFile("temp/schema2.sql", schema2)

	// Print diff
	printDiff(schema1, schema2, args[1], args[2])
}
