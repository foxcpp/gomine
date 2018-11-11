package gomine

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func findSystemJava() (string, error) {
	javaBinExt := ""
	if runtime.GOOS == "windows" {
		javaBinExt = ".exe"
	}

	jreHome := os.Getenv("JRE_HOME")
	if jreHome != "" {
		return filepath.Join(jreHome, "bin", "java" + javaBinExt), nil
	}

	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		return filepath.Join(javaHome, "bin", "java" + javaBinExt), nil
	}

	path, err := exec.LookPath("java" + javaBinExt)
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return path, err
	}
	return absPath, nil
}