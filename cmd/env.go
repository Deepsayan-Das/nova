package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// secretPatterns contains case-insensitive substrings that indicate a key holds a secret value.
// If a key name contains any of these, its value is masked in list output.
var secretPatterns = []string{"SECRET", "PASSWORD", "KEY", "TOKEN", "API"}

// envCmd is the parent command for environment-variable management subcommands.
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage project environment variables",
	Long: `env provides subcommands to list, set, get, and audit environment
variables stored in a project's .env file. It works on any project directory,
not just Nova-scaffolded ones.`,
}

// --- nova env list [path] ---------------------------------------------------

var envListCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List all environment variables defined in .env",
	Long: `list reads the .env file (and .env.example, if present) in the target
directory and prints every key with a colored status indicator.

Values for keys whose name matches common secret patterns (SECRET, PASSWORD,
KEY, TOKEN, API) are masked as **** to avoid accidental leakage.`,
	Run: runEnvList,
}

func runEnvList(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	dir := resolveDir(args)

	envPath := filepath.Join(dir, ".env")
	examplePath := filepath.Join(dir, ".env.example")

	envVars, err := parseEnvFile(envPath)
	if err != nil && !os.IsNotExist(err) {
		red.Printf("Error reading .env: %v\n", err)
		os.Exit(1)
	}

	exampleVars, _ := parseEnvFile(examplePath)

	// Merge keys: everything in .env plus anything only in .env.example
	allKeys := mergeKeyOrder(envVars, exampleVars)

	if len(allKeys) == 0 {
		cyan.Println("No environment variables found (no .env or .env.example in this directory).")
		return
	}

	cyan.Println("========== Environment Variables ==========")
	for _, key := range allKeys {
		val, set := envVars[key]
		if set {
			display := val
			if isSecret(key) {
				display = "****"
			}
			green.Printf("  ✓ %s = %s\n", key, display)
		} else {
			red.Printf("  ✗ %s (defined in .env.example but not set)\n", key)
		}
	}
	cyan.Println("===========================================")
}

// --- nova env set KEY=VALUE [path] -------------------------------------------

var envSetCmd = &cobra.Command{
	Use:   "set KEY=VALUE [path]",
	Short: "Set or update an environment variable in .env",
	Long: `set writes a key-value pair into the project's .env file. If the file
does not exist it is created. Existing keys are updated in-place; new keys are
appended. All other lines are preserved exactly as-is.`,
	Run: runEnvSet,
}

func runEnvSet(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	if len(args) < 1 {
		red.Println("Error: expected KEY=VALUE argument")
		fmt.Println("Usage: nova env set KEY=VALUE [path]")
		os.Exit(1)
	}

	kv := args[0]
	eqIdx := strings.Index(kv, "=")
	if eqIdx <= 0 {
		red.Println("Error: argument must be in KEY=VALUE format")
		os.Exit(1)
	}
	key := kv[:eqIdx]
	value := kv[eqIdx+1:]

	dir := "."
	if len(args) > 1 {
		dir = args[1]
	}

	envPath := filepath.Join(dir, ".env")

	// Read existing lines (or start empty)
	var lines []string
	if data, err := os.ReadFile(envPath); err == nil {
		lines = strings.Split(string(data), "\n")
	}

	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if eqPos := strings.Index(trimmed, "="); eqPos > 0 {
			lineKey := trimmed[:eqPos]
			if lineKey == key {
				lines[i] = key + "=" + value
				found = true
				break
			}
		}
	}

	if !found {
		// Append — make sure there's a trailing newline before appending
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = append(lines[:len(lines)-1], key+"="+value, "")
		} else {
			lines = append(lines, key+"="+value)
		}
	}

	if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		red.Printf("Error writing .env: %v\n", err)
		os.Exit(1)
	}

	green.Printf("✓ %s set in %s\n", key, envPath)
}

// --- nova env get KEY [path] -------------------------------------------------

var envGetCmd = &cobra.Command{
	Use:   "get KEY [path]",
	Short: "Print the value of a single environment variable from .env",
	Long: `get reads the value of a specific key from the project's .env file and
prints it. Unlike list, no masking is applied — you asked for the key explicitly.`,
	Run: runEnvGet,
}

func runEnvGet(cmd *cobra.Command, args []string) {
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)

	if len(args) < 1 {
		red.Println("Error: expected KEY argument")
		fmt.Println("Usage: nova env get KEY [path]")
		os.Exit(1)
	}

	key := args[0]
	dir := "."
	if len(args) > 1 {
		dir = args[1]
	}

	envPath := filepath.Join(dir, ".env")
	envVars, err := parseEnvFile(envPath)
	if err != nil {
		red.Printf("Error reading .env: %v\n", err)
		os.Exit(1)
	}

	val, ok := envVars[key]
	if !ok {
		yellow.Printf("%s is not set in %s\n", key, envPath)
		os.Exit(1)
	}

	fmt.Println(val)
}

// --- nova env check [path] ---------------------------------------------------

var envCheckCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Audit .env against .env.example for missing or extra keys",
	Long: `check compares the keys present in .env against those declared in
.env.example. Missing keys are reported as warnings; keys present in .env but
not documented in .env.example are reported as informational.`,
	Run: runEnvCheck,
}

func runEnvCheck(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)

	dir := resolveDir(args)

	envPath := filepath.Join(dir, ".env")
	examplePath := filepath.Join(dir, ".env.example")

	envVars, err := parseEnvFile(envPath)
	if err != nil && !os.IsNotExist(err) {
		red.Printf("Error reading .env: %v\n", err)
		os.Exit(1)
	}

	exampleVars, err := parseEnvFile(examplePath)
	if err != nil {
		yellow.Println("No .env.example found — nothing to check against.")
		return
	}

	cyan.Println("========== Environment Check ==========")

	missing := 0
	extra := 0

	// Keys expected (.env.example) but missing from .env
	for key := range exampleVars {
		if _, ok := envVars[key]; !ok {
			yellow.Printf("  ⚠ MISSING: %s (defined in .env.example but not in .env)\n", key)
			missing++
		}
	}

	// Keys in .env but not documented in .env.example
	for key := range envVars {
		if _, ok := exampleVars[key]; !ok {
			cyan.Printf("  ℹ EXTRA: %s (in .env but not in .env.example)\n", key)
			extra++
		}
	}

	if missing == 0 && extra == 0 {
		green.Println("  ✓ .env is fully in sync with .env.example")
	}

	fmt.Println("---------------------------------------")
	if missing > 0 {
		yellow.Printf("Result: %d missing, %d undocumented\n", missing, extra)
	} else {
		green.Printf("Result: %d missing, %d undocumented\n", missing, extra)
	}
	cyan.Println("=======================================")
}

// --- helpers -----------------------------------------------------------------

// resolveDir returns the directory from the first arg, defaulting to "./"
func resolveDir(args []string) string {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return args[0]
	}
	return "./"
}

// parseEnvFile reads a .env-format file and returns a map of KEY → VALUE.
// Blank lines and lines starting with # are skipped.
func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx <= 0 {
			continue
		}
		key := line[:eqIdx]
		value := line[eqIdx+1:]
		// Strip optional surrounding quotes
		value = strings.Trim(value, `"'`)
		vars[key] = value
	}
	return vars, scanner.Err()
}

// isSecret returns true if the key name matches any common secret pattern.
func isSecret(key string) bool {
	upper := strings.ToUpper(key)
	for _, pat := range secretPatterns {
		if strings.Contains(upper, pat) {
			return true
		}
	}
	return false
}

// mergeKeyOrder returns a deduplicated ordered list of keys from both maps,
// with envVars keys first (preserving insertion order is not guaranteed by Go
// maps, so this is "all env keys" then "example-only keys").
func mergeKeyOrder(envVars, exampleVars map[string]string) []string {
	seen := make(map[string]bool)
	var keys []string

	for k := range envVars {
		if !seen[k] {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	for k := range exampleVars {
		if !seen[k] {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	return keys
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envCheckCmd)
}
