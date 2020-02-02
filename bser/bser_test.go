package bser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting wd %s\n", err)
		os.Exit(1)
	}
	dir := filepath.Join(wd, "test")

	if err := os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) {
		fmt.Printf("Error creating testdir %s\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("npm", "install", "bser")
	cmd.Dir = dir
	if _, err := cmd.Output(); err != nil {
		fmt.Printf("Error installing node bser %s\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}
