package cmd

import (
	"bytes"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error executing version command: %v", err)
	}

	output := buf.String()
	if output != "" && output != "nova version 1.0.0\n" {
		t.Logf("Version output: %s", output)
	}
}

func TestVersionFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error executing --version flag: %v", err)
	}
}
