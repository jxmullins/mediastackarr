package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseEnvFile reads a .env file and returns a map of key-value pairs
func ParseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	env := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Find the first = sign
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Remove surrounding quotes if present
		value = trimQuotes(value)

		// Remove inline comments (but be careful with # in quoted strings)
		value = removeInlineComment(value)

		// Expand ${VAR} references
		value = expandVariables(value, env)

		env[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}

	return env, nil
}

// trimQuotes removes surrounding single or double quotes
func trimQuotes(s string) string {
	if len(s) < 2 {
		return s
	}

	if (s[0] == '"' && s[len(s)-1] == '"') ||
		(s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}

	return s
}

// removeInlineComment removes comments from the end of a line
// but preserves # characters that appear in values
func removeInlineComment(s string) string {
	// Simple approach: find # that's preceded by whitespace
	// This won't handle all edge cases but works for most .env files
	for i := 1; i < len(s); i++ {
		if s[i] == '#' && (s[i-1] == ' ' || s[i-1] == '\t') {
			return strings.TrimSpace(s[:i-1])
		}
	}
	return s
}

// expandVariables replaces ${VAR} with the value of VAR from the env map
func expandVariables(s string, env map[string]string) string {
	result := s

	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		varName := result[start+2 : end]
		varValue := ""

		// Check if there's a default value (${VAR:-default})
		if idx := strings.Index(varName, ":-"); idx != -1 {
			defaultVal := varName[idx+2:]
			varName = varName[:idx]
			if val, ok := env[varName]; ok && val != "" {
				varValue = val
			} else {
				varValue = defaultVal
			}
		} else if val, ok := env[varName]; ok {
			varValue = val
		} else if val, ok := os.LookupEnv(varName); ok {
			varValue = val
		}

		result = result[:start] + varValue + result[end+1:]
	}

	return result
}

// ExportToEnvironment exports all env vars to the current process
func ExportToEnvironment(env map[string]string) {
	for key, value := range env {
		os.Setenv(key, value)
	}
}

// ValidateRequiredVars checks that all required variables are present
func ValidateRequiredVars(env map[string]string, required []string) []string {
	missing := []string{}
	for _, key := range required {
		if val, ok := env[key]; !ok || val == "" {
			missing = append(missing, key)
		}
	}
	return missing
}
