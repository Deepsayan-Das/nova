package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type NovaMetadata struct {
	Language    string `json:"language" yaml:"language"`
	Template    string `json:"template" yaml:"template"`
	CreatedAt   string `json:"created" yaml:"created"`
	NovaVersion string `json:"nova_version" yaml:"nova_version"`
}

var projectDoctorCmd = &cobra.Command{
	Use:   "doctor [path]",
	Short: "Check the health of the current project",
	Long: `project doctor reads this project's .nova.yaml metadata (written by
'nova init' / 'nova project init') to determine its language and template, then runs
checks relevant to that specific project — e.g. whether dependencies are
installed, config files are valid, and required tools exist.`,
	Run: runProjectDoctor,
}

func init() {
	projectCmd.AddCommand(projectDoctorCmd)
}

func runProjectDoctor(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)

	var dest string
	if len(args) <= 0 {
		dest = "./"
	} else {
		dest = args[0]
	}

	info, err := os.Stat(dest)
	if err != nil || !info.IsDir() {
		red.Printf("Error: Destination directory '%s' does not exist or is not a directory\n", dest)
		os.Exit(1)
	}

	metadataPath := filepath.Join(dest, ".nova.yaml")
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		red.Printf("Error: Could not read metadata file '%s' (is this a Nova project?)\n", metadataPath)
		os.Exit(1)
	}

	var metadataYaml NovaMetadata
	if err := yaml.Unmarshal(metadataBytes, &metadataYaml); err != nil {
		red.Printf("Error: Could not parse metadata file: %v\n", err)
		os.Exit(1)
	}

	cyan.Println("========== Running Project Health Checks ==========")
	fmt.Printf("Project Path : %s\n", dest)
	fmt.Printf("Language     : %s\n", metadataYaml.Language)
	fmt.Printf("Template     : %s\n", metadataYaml.Template)
	fmt.Printf("Created At   : %s\n", metadataYaml.CreatedAt)
	fmt.Printf("Nova Version : %s\n", metadataYaml.NovaVersion)
	fmt.Println("---------------------------------------------------")

	passed := 0
	warnings := 0
	failed := 0

	// 1. General Checks
	green.Println("  [SUCCESS] .nova.yaml metadata valid")
	passed++

	gitDir := filepath.Join(dest, ".git")
	if dirExists(gitDir) {
		green.Println("  [SUCCESS] Git repository initialized (.git)")
		passed++
	} else {
		yellow.Println("  [WARNING] Git repository not initialized (run 'git init')")
		warnings++
	}

	// 2. Language-Specific Checks
	lang := strings.ToLower(strings.TrimSpace(metadataYaml.Language))
	cyan.Printf("\nChecking %s project requirements:\n", lang)

	switch lang {
	case "go", "golang":
		// Check go compiler
		if ver, ok := getToolVersion("go", "version"); ok {
			green.Printf("  [SUCCESS] Go runtime: %s\n", ver)
			passed++
		} else {
			red.Println("  [FAILED] Go compiler ('go') not found in PATH")
			failed++
		}

		// Check go.mod
		goModPath := filepath.Join(dest, "go.mod")
		if fileExists(goModPath) {
			green.Println("  [SUCCESS] go.mod file found")
			passed++

			// Run go mod verify
			vCmd := exec.Command("go", "mod", "verify")
			vCmd.Dir = dest
			if err := vCmd.Run(); err == nil {
				green.Println("  [SUCCESS] Go module dependencies verified (go mod verify)")
				passed++
			} else {
				yellow.Println("  [WARNING] Go module verification warning (run 'go mod tidy')")
				warnings++
			}
		} else {
			red.Println("  [FAILED] go.mod missing in project root")
			failed++
		}

	case "node", "javascript", "js", "typescript", "ts", "next.js", "next", "vite":
		// Check node runtime
		if ver, ok := getToolVersion("node", "-v"); ok {
			green.Printf("  [SUCCESS] Node.js runtime: %s\n", ver)
			passed++
		} else {
			red.Println("  [FAILED] Node.js ('node') not found in PATH")
			failed++
		}

		// Check package manager
		var pkgMgr string
		for _, pm := range []string{"npm", "pnpm", "yarn", "bun"} {
			if _, ok := getToolVersion(pm, "-v"); ok {
				pkgMgr = pm
				break
			}
		}
		if pkgMgr != "" {
			green.Printf("  [SUCCESS] Package manager found: %s\n", pkgMgr)
			passed++
		} else {
			yellow.Println("  [WARNING] No Node package manager (npm/yarn/pnpm/bun) found in PATH")
			warnings++
		}

		// Check package.json
		pkgJsonPath := filepath.Join(dest, "package.json")
		if fileExists(pkgJsonPath) {
			green.Println("  [SUCCESS] package.json found")
			passed++
		} else {
			red.Println("  [FAILED] package.json missing in project root")
			failed++
		}

		// Check node_modules
		nodeModulesPath := filepath.Join(dest, "node_modules")
		if dirExists(nodeModulesPath) {
			green.Println("  [SUCCESS] node_modules directory present")
			passed++
		} else {
			yellow.Println("  [WARNING] node_modules missing (run 'npm install')")
			warnings++
		}

	case "python", "python3", "py":
		// Check python interpreter
		pyBin := "python3"
		ver, ok := getToolVersion("python3", "--version")
		if !ok {
			pyBin = "python"
			ver, ok = getToolVersion("python", "--version")
		}
		if ok {
			green.Printf("  [SUCCESS] Python interpreter (%s): %s\n", pyBin, ver)
			passed++
		} else {
			red.Println("  [FAILED] Python interpreter ('python3' / 'python') not found in PATH")
			failed++
		}

		// Check dependency manifest
		var manifest string
		for _, m := range []string{"requirements.txt", "pyproject.toml", "Pipfile", "setup.py"} {
			if fileExists(filepath.Join(dest, m)) {
				manifest = m
				break
			}
		}
		if manifest != "" {
			green.Printf("  [SUCCESS] Dependency manifest found (%s)\n", manifest)
			passed++
		} else {
			yellow.Println("  [WARNING] No Python dependency manifest found (requirements.txt / pyproject.toml)")
			warnings++
		}

		// Check virtual environment
		var venvDir string
		for _, v := range []string{".venv", "venv", "env"} {
			if dirExists(filepath.Join(dest, v)) {
				venvDir = v
				break
			}
		}
		if venvDir != "" {
			green.Printf("  [SUCCESS] Virtual environment present (%s/)\n", venvDir)
			passed++
		} else {
			yellow.Println("  [WARNING] Virtual environment missing (consider creating .venv)")
			warnings++
		}

	case "java":
		// Check java & javac
		if ver, ok := getToolVersion("java", "-version"); ok {
			green.Printf("  [SUCCESS] Java runtime: %s\n", ver)
			passed++
		} else {
			red.Println("  [FAILED] Java runtime ('java') not found in PATH")
			failed++
		}

		if _, ok := getToolVersion("javac", "-version"); ok {
			green.Println("  [SUCCESS] Java compiler ('javac') found in PATH")
			passed++
		} else {
			yellow.Println("  [WARNING] Java compiler ('javac') not found in PATH")
			warnings++
		}

		// Check build file
		var buildFile string
		for _, bf := range []string{"pom.xml", "build.gradle", "build.gradle.kts"} {
			if fileExists(filepath.Join(dest, bf)) {
				buildFile = bf
				break
			}
		}
		if buildFile != "" {
			green.Printf("  [SUCCESS] Java build file found (%s)\n", buildFile)
			passed++
		} else {
			yellow.Println("  [WARNING] No Java build file found (pom.xml / build.gradle)")
			warnings++
		}

	case "c", "c++", "cpp":
		// Check compiler
		var compiler string
		for _, comp := range []string{"gcc", "g++", "clang", "clang++"} {
			if _, ok := getToolVersion(comp, "--version"); ok {
				compiler = comp
				break
			}
		}
		if compiler != "" {
			green.Printf("  [SUCCESS] C/C++ compiler found (%s)\n", compiler)
			passed++
		} else {
			red.Println("  [FAILED] No C/C++ compiler (gcc/g++/clang) found in PATH")
			failed++
		}

		// Check build tool
		var buildTool string
		for _, bt := range []string{"make", "cmake", "ninja"} {
			if _, ok := getToolVersion(bt, "--version"); ok {
				buildTool = bt
				break
			}
		}
		if buildTool != "" {
			green.Printf("  [SUCCESS] Build tool found (%s)\n", buildTool)
			passed++
		} else {
			yellow.Println("  [WARNING] No build tool (make/cmake/ninja) found in PATH")
			warnings++
		}

		// Check Makefile / CMakeLists.txt
		var configFile string
		for _, cf := range []string{"Makefile", "makefile", "CMakeLists.txt"} {
			if fileExists(filepath.Join(dest, cf)) {
				configFile = cf
				break
			}
		}
		if configFile != "" {
			green.Printf("  [SUCCESS] Build configuration found (%s)\n", configFile)
			passed++
		} else {
			yellow.Println("  [WARNING] No build configuration file found (Makefile / CMakeLists.txt)")
			warnings++
		}

	default:
		cyan.Printf("  No extra checks registered for language '%s'\n", lang)
	}

	fmt.Println("\n---------------------------------------------------")
	if failed > 0 {
		red.Printf("Project Doctor finished: %d passed, %d warning(s), %d failed\n", passed, warnings, failed)
	} else if warnings > 0 {
		yellow.Printf("Project Doctor finished: %d passed, %d warning(s), %d failed\n", passed, warnings, failed)
	} else {
		green.Printf("Project Doctor finished: %d passed, %d warning(s), %d failed\n", passed, warnings, failed)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func getToolVersion(name string, flag string) (string, bool) {
	_, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	out, err := exec.Command(name, flag).CombinedOutput()
	if err != nil && len(out) == 0 {
		return "", false
	}
	str := strings.TrimSpace(string(out))
	lines := strings.Split(str, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), true
	}
	return "", true
}
