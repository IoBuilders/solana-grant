package main

import (
	"bytes"
	"dlt-ingress/src/cmd/core"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

var tmpl *template.Template
var completedOk = false

func init() {
	var err error
	tmpl, err = template.ParseFiles("./../cmd/gencallbacks/method.tmpl")
	if err != nil {
		fmt.Println("❌ Failed to load template:", err)
		os.Exit(1)
	}
}

func starting() {
	fmt.Println("──────────────────────────────────────")
	fmt.Println("🔄  Starting callbacks generator")
	fmt.Println("──────────────────────────────────────")
}

func ending() {
	fmt.Println("───────────────────────────────")

	if completedOk {
		fmt.Println("✅ Callbacks generation finished successfully")
	} else {
		fmt.Println("❌ Callbacks generation failed")
	}
	fmt.Println("───────────────────────────────")
	fmt.Println()
	fmt.Println()
}

func main() {
	starting()
	defer ending()
	root := "./../main/"
	appServiceRelativePath := "/internal/app/worker/process/callback"
	boundedContexts := cmdcore.BOUNDED_CONTEXTS
	for _, boundedContext := range boundedContexts {
		if cmdcore.SkipCallback(boundedContext) {
			fmt.Printf("-- Skipping %s: Callbacks not needed\n", boundedContext)
			continue
		}
		if !cmdcore.IsMigrated(boundedContext) {
			fmt.Printf("-- Skipping %s: Handled by old generator\n", boundedContext)
			continue
		}
		err := generateCode(filepath.Join(root, boundedContext, appServiceRelativePath), "callback_process_worker_gen.go")
		if err != nil {
			fmt.Printf("❌ Failed to generate callback worker for %s: %v\n", boundedContext, err)
			os.Exit(1)
		}
	}
	completedOk = true
}

func generateCode(origPath, fileName string) error {
	genPath := filepath.Join(origPath, fileName)
	fmt.Println("⚙️  Generating:", genPath)

	err := os.MkdirAll(origPath, os.ModePerm)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct{}{})
	if err != nil {
		return err
	}

	return os.WriteFile(genPath, buf.Bytes(), 0644)
}
