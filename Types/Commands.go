package Types

type TestItem struct {
	Name    string   `json:"name" yaml:"name"`
	Command []string `json:"command" yaml:"command"`
}

type ToolCheck struct {
	Name        string     `json:"name" yaml:"name"`
	Key         string     `json:"key" yaml:"key"`
	Description string     `json:"description" yaml:"description"`
	Executables []string   `json:"executables,omitempty" yaml:"executables,omitempty"`
	Tests       []TestItem `json:"tests,omitempty" yaml:"tests,omitempty"`
}

type DoctorConfig struct {
	Checks []ToolCheck `json:"checks" yaml:"checks"`
}

type Commands struct {
	Name        string   `json:"name" yaml:"name"`
	Executables []string `json:"executables" yaml:"executables"`
}
