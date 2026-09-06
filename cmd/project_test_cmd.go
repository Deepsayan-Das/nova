package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var projectTestCmd = &cobra.Command{
	Use:   "test [path]",
	Short: "Run tests for the current project",
	Long: `project test reads this project's .nova.yaml metadata (written by
'nova init' / 'nova project init') to determine its language and template, then runs
the appropriate test command for the project stack.`,
	Run: runProjectTest,
}

func init() {
	projectCmd.AddCommand(projectTestCmd)
}

func runProjectTest(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	var dest string
	var extraArgs []string

	if len(args) > 0 {
		dest = args[0]
		extraArgs = args[1:]
	} else {
		dest = "./"
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

	lang := strings.ToLower(strings.TrimSpace(metadataYaml.Language))

	execName, execArgs, err := resolveTestRunner(lang, dest, extraArgs)
	if err != nil {
		red.Println(err.Error())
		os.Exit(1)
	}

	cyan.Println("========== Running Project Tests ==========")
	fmt.Printf("Project Path : %s\n", dest)
	fmt.Printf("Language     : %s\n", metadataYaml.Language)
	fmt.Printf("Template     : %s\n", metadataYaml.Template)
	fmt.Printf("Test Runner  : %s %s\n", execName, strings.Join(execArgs, " "))
	fmt.Println("-------------------------------------------")

	runCmd := exec.Command(execName, execArgs...)
	runCmd.Dir = dest
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	runCmd.Stdin = os.Stdin

	if err := runCmd.Run(); err != nil {
		fmt.Println("-------------------------------------------")
		red.Printf("[FAILED] Test suite failed for %s project: %v\n", metadataYaml.Language, err)
		os.Exit(1)
	}

	fmt.Println("-------------------------------------------")
	green.Printf("[SUCCESS] All tests passed successfully for %s project!\n", metadataYaml.Language)
}

func resolveTestRunner(lang string, dest string, extraArgs []string) (string, []string, error) {
	var execName string
	var execArgs []string

	switch lang {
	case "go", "golang":
		execName = "go"
		execArgs = append([]string{"test", "./..."}, extraArgs...)

	case "node", "javascript", "js", "typescript", "ts", "next.js", "next", "vite":
		var pkgMgr string
		for _, pm := range []string{"npm", "pnpm", "yarn", "bun"} {
			if _, err := exec.LookPath(pm); err == nil {
				pkgMgr = pm
				break
			}
		}
		if pkgMgr == "" {
			return "", nil, fmt.Errorf("Error: No Node package manager (npm/yarn/pnpm/bun) found in PATH")
		}
		execName = pkgMgr
		execArgs = append([]string{"test"}, extraArgs...)

	case "python", "python3", "py":
		if _, err := exec.LookPath("pytest"); err == nil {
			execName = "pytest"
			execArgs = extraArgs
		} else if _, err := exec.LookPath("python3"); err == nil {
			execName = "python3"
			execArgs = append([]string{"-m", "unittest", "discover"}, extraArgs...)
		} else if _, err := exec.LookPath("python"); err == nil {
			execName = "python"
			execArgs = append([]string{"-m", "unittest", "discover"}, extraArgs...)
		} else {
			return "", nil, fmt.Errorf("Error: No Python interpreter or pytest found in PATH")
		}

	case "java":
		if fileExists(filepath.Join(dest, "gradlew")) || fileExists(filepath.Join(dest, "gradlew.bat")) {
			if fileExists(filepath.Join(dest, "gradlew.bat")) && runtime.GOOS == "windows" {
				execName = "cmd.exe"
				execArgs = append([]string{"/c", "gradlew.bat", "test"}, extraArgs...)
			} else {
				gradPath := filepath.Join(dest, "gradlew")
				if abs, err := filepath.Abs(gradPath); err == nil {
					execName = abs
				} else {
					execName = gradPath
				}
				execArgs = append([]string{"test"}, extraArgs...)
			}
		} else if fileExists(filepath.Join(dest, "pom.xml")) {
			execName = "mvn"
			execArgs = append([]string{"test"}, extraArgs...)
		} else if fileExists(filepath.Join(dest, "build.gradle")) || fileExists(filepath.Join(dest, "build.gradle.kts")) {
			execName = "gradle"
			execArgs = append([]string{"test"}, extraArgs...)
		} else {
			color.Yellow("Warning: No standard Java build file (pom.xml / build.gradle) found.")
			execName = "mvn"
			execArgs = append([]string{"test"}, extraArgs...)
		}

	case "c", "c++", "cpp":
		if fileExists(filepath.Join(dest, "Makefile")) || fileExists(filepath.Join(dest, "makefile")) {
			execName = "make"
			execArgs = append([]string{"test"}, extraArgs...)
		} else if fileExists(filepath.Join(dest, "CMakeLists.txt")) {
			execName = "ctest"
			execArgs = extraArgs
		} else {
			execName = "make"
			execArgs = append([]string{"test"}, extraArgs...)
		}

	default:
		return "", nil, fmt.Errorf("Error: Unsupported language for testing: '%s'", lang)
	}

	return execName, execArgs, nil
}
