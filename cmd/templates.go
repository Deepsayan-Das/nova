package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed templates.yaml
var defaultTemplatesYAML []byte

type ProjectFile struct {
	Path    string `json:"path" yaml:"path"`
	Content string `json:"content" yaml:"content"`
}

type Template struct {
	Language string        `json:"language" yaml:"language"`
	Name     string        `json:"name" yaml:"name"`
	Files    []ProjectFile `json:"files,omitempty" yaml:"files,omitempty"`
	Dirs     []string      `json:"dirs,omitempty" yaml:"dirs,omitempty"`
}

type TemplatesConfig struct {
	Templates []Template `json:"templates" yaml:"templates"`
}

func GetAllTemplates() []Template {
	var yamlData []byte

	if homeDir, err := os.UserHomeDir(); err == nil {
		userConfigPath := filepath.Join(homeDir, ".config", "nova", "templates.yaml")
		if data, err := os.ReadFile(userConfigPath); err == nil {
			yamlData = data
		}
	}

	if len(yamlData) == 0 {
		yamlData = defaultTemplatesYAML
	}

	var config TemplatesConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		fmt.Printf("Error parsing templates.yaml: %v\n", err)
		return nil
	}
	return config.Templates
}

func matchLanguage(reqLang, tmplLang string) bool {
	req := strings.ToLower(strings.TrimSpace(reqLang))
	tmpl := strings.ToLower(strings.TrimSpace(tmplLang))

	if req == tmpl {
		return true
	}
	switch req {
	case "js", "javascript", "node", "nodejs":
		return tmpl == "node" || tmpl == "js"
	case "next", "nextjs", "next.js":
		return tmpl == "next.js" || tmpl == "nextjs" || tmpl == "next"
	case "cpp", "c++":
		return tmpl == "c++" || tmpl == "cpp"
	case "py", "python":
		return tmpl == "python" || tmpl == "py"
	}
	return false
}

func GetTemplate(lang, templateName string) *Template {
	for _, t := range GetAllTemplates() {
		if matchLanguage(lang, t.Language) && strings.EqualFold(t.Name, templateName) {
			return &t
		}
	}
	return nil
}
