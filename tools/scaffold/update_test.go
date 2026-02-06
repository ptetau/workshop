package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestUpdateConcept(t *testing.T) {
	workDir, err := os.MkdirTemp("", "update_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	// 1. Initial Scaffold
	runInit([]string{
		"--root", workDir,
		"--module", "testapp",
		"--concept", "User",
		"--field", "User:Name:string",
	})

	modelPath := filepath.Join(workDir, "internal/domain/user/model.go")

	// 2. Add Custom Method
	content, _ := os.ReadFile(modelPath)
	customMethod := "\nfunc (u *User) SayHello() string { return \"Hello\" }\n"
	os.WriteFile(modelPath, append(content, []byte(customMethod)...), 0644)

	// 3. Update Scaffold (Add Age)
	runInit([]string{
		"--root", workDir,
		"--module", "testapp",
		"--concept", "User",
		"--field", "User:Name:string",
		"--field", "User:Age:int",
	})

	// 4. Verify
	newContent, _ := os.ReadFile(modelPath)
	sContent := string(newContent)

	if matched, _ := regexp.MatchString(`Age\s+int`, sContent); !matched {
		t.Errorf("Missing new field Age. Content:\n%s", sContent)
	}
	if !strings.Contains(sContent, "SayHello") {
		t.Errorf("Lost custom method SayHello. Content:\n%s", sContent)
	}
}

func TestUpdateOrchestratorUI(t *testing.T) {
	workDir, err := os.MkdirTemp("", "update_ui_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	// 1. Initial
	runInit([]string{
		"--root", workDir,
		"--module", "testapp",
		"--orchestrator", "Login",
		"--param", "Login:User:string",
		"--route", "POST:/login:Login",
	})

	// 2. Update (Add Password)
	runInit([]string{
		"--root", workDir,
		"--module", "testapp",
		"--orchestrator", "Login",
		"--param", "Login:User:string",
		"--param", "Login:Pass:string",
		"--route", "POST:/login:Login",
	})

	// 3. Verify HTML
	htmlPath := filepath.Join(workDir, "internal/adapters/http/templates/form_login.html")
	content, _ := os.ReadFile(htmlPath)
	sContent := string(content)

	if !strings.Contains(sContent, "name=\"Pass\"") {
		t.Error("Missing new input for Pass")
	}
}
