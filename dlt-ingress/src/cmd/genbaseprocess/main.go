package main

import (
	"bytes"
	"dlt-ingress/src/cmd/core"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

var adapterTmpl *template.Template
var portTmpl *template.Template
var handlerTmpl *template.Template
var queryTmpl *template.Template
var completedOk = false

func init() {
	var err error
	tmplRoot := "./../cmd/genbaseprocess/"
	adapterTmpl, err = template.ParseFiles(tmplRoot + "adapter.tmpl")
	portTmpl, err = template.ParseFiles(tmplRoot + "port.tmpl")
	handlerTmpl, err = template.ParseFiles(tmplRoot + "handler.tmpl")
	queryTmpl, err = template.ParseFiles(tmplRoot + "query.tmpl")
	if err != nil {
		fmt.Println("❌ Failed to load template:", err)
		os.Exit(1)
	}
}

func starting() {
	fmt.Println("──────────────────────────────────────")
	fmt.Println("🧩  Starting baseprocess generator")
	fmt.Println("──────────────────────────────────────")
}

func ending() {
	fmt.Println("───────────────────────────────")

	if completedOk {
		fmt.Println("✅ Baseprocess generation finished successfully")
	} else {
		fmt.Println("❌ Baseprocess generation failed")
	}
	fmt.Println("───────────────────────────────")
	fmt.Println()
	fmt.Println()
}

func main() {
	starting()
	defer ending()

	root := "./../main/"
	boundedContexts := []string{
		//cmdcore.ISSUANCE,
		cmdcore.REGISTRY,
		//cmdcore.ASSETLIFECYCLE,
		cmdcore.SHARED,
		cmdcore.PRIMARY_MARKET,
		cmdcore.SETTLEMENT,
		cmdcore.COMPLIANCE,
	}
	apiRelativePath := "/port/api/process/"
	packageName := "getprocess"
	filePrefix := "get_process"
	queryRelativePath := "/app/query/"
	for _, boundedContext := range boundedContexts {
		err := generateCode(root, boundedContext, apiRelativePath, filePrefix+"_http_adapter", packageName, adapterTmpl)
		if err != nil {
			fmt.Println(fmt.Sprintf("❌ Error generating adapter file for bounded context (%s): %s\n", boundedContext, err))
			os.Exit(1)
		}
		err = generateCode(root, boundedContext, apiRelativePath, filePrefix+"_port", packageName, portTmpl)
		if err != nil {
			fmt.Println(fmt.Sprintf("❌ Error generating port file for bounded context (%s): %s\n", boundedContext, err))
			os.Exit(1)
		}
		err = generateCode(root, boundedContext, queryRelativePath, filePrefix+"_handler", packageName, handlerTmpl)
		if err != nil {
			fmt.Println(fmt.Sprintf("❌ Error generating handler file for bounded context (%s): %s\n", boundedContext, err))
			os.Exit(1)
		}
		err = generateCode(root, boundedContext, queryRelativePath, filePrefix+"_query", packageName, queryTmpl)
		if err != nil {
			fmt.Println(fmt.Sprintf("❌ Error generating query file for bounded context (%s): %s\n", boundedContext, err))
			os.Exit(1)
		}
	}

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error executing go fmt: %v", err)
	}

	completedOk = true
}

func generateCode(root, boundedContext, relativePath, fileName, packageName string, tmpl *template.Template) error {
	genDir := filepath.Join(filepath.Dir(root), boundedContext, relativePath, packageName)
	genPath := filepath.Join(genDir, fileName+"_gen.go")
	fmt.Println("⚙️  Generating:", genPath)

	if err := os.MkdirAll(genDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", genDir, err)
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, struct {
		Package        string
		BoundedContext string
	}{
		Package:        packageName,
		BoundedContext: boundedContext,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(genPath, buf.Bytes(), 0644)
}
