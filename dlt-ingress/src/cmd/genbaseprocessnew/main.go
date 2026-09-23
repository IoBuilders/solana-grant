package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	queryTmpl   *template.Template
	handlerTmpl *template.Template
	portTmpl    *template.Template
	adapterTmpl *template.Template
)

var completedOk = false

func init() {
	tmplRoot := "./../cmd/genbaseprocessnew/"

	queryTmpl = mustLoadTemplate("query.tmpl", tmplRoot)
	handlerTmpl = mustLoadTemplate("handler.tmpl", tmplRoot)
	portTmpl = mustLoadTemplate("port.tmpl", tmplRoot)
	adapterTmpl = mustLoadTemplate("adapter.tmpl", tmplRoot)
}

func mustLoadTemplate(name, dir string) *template.Template {
	tmpl, err := template.New(name).ParseFiles(filepath.Join(dir, name))
	if err != nil {
		panic(fmt.Sprintf("failed to load template %s: %v", name, err))
	}
	return tmpl
}

func starting() {
	fmt.Println("──────────────────────────────────────")
	fmt.Println("🧩  Starting genbaseprocessnew generator")
	fmt.Println("──────────────────────────────────────")
}

func ending() {
	fmt.Println("───────────────────────────────")

	if completedOk {
		fmt.Println("✅ genbaseprocessnew generation finished successfully")
	} else {
		fmt.Println("❌ genbaseprocessnew generation failed")
	}
	fmt.Println("───────────────────────────────")
	fmt.Println()
	fmt.Println()
}

func main() {
	starting()
	defer ending()

	root := "./../main/"

	queryDir := filepath.Join(root, "issuance/internal/app/query/getprocess")
	apiDir := filepath.Join(root, "issuance/internal/infra/api/process/getprocess")

	// query.tmpl bakes in its own QueryName()/ResponseType() (the same hardcoded-path
	// approach genqueryname itself generates) instead of relying on that companion
	// generator, so get_process_query_gen.go can safely carry the _gen.go suffix like
	// its siblings — genqueryname's file walk skips any *_gen.go file outright.
	files := []struct {
		path string
		tmpl *template.Template
	}{
		{filepath.Join(queryDir, "get_process_query_gen.go"), queryTmpl},
		{filepath.Join(queryDir, "get_process_handler_gen.go"), handlerTmpl},
		{filepath.Join(apiDir, "get_process_port_gen.go"), portTmpl},
		{filepath.Join(apiDir, "get_process_http_adapter_gen.go"), adapterTmpl},
	}
	for _, f := range files {
		if err := generateFile(f.path, f.tmpl); err != nil {
			fmt.Println("❌ Error generating file", f.path, ":", err)
			os.Exit(1)
		}
	}

	injectQueryHandler(root)
	injectRouterRegistration(root)

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("❌ Error executing go fmt:", err)
		os.Exit(1)
	}

	completedOk = true
}

func generateFile(path string, tmpl *template.Template) error {
	fmt.Println("⚙️  Generating:", path)

	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(path), err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// injectQueryHandler wires getprocessquery.NewHandler into config/queryhandlers.go's
// SetupQueryHandlers, so the handler registers with the query bus.
func injectQueryHandler(root string) {
	filePath := filepath.Join(root, "issuance/config/queryhandlers.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ Error reading", filePath, ":", err)
		os.Exit(1)
	}
	sContent := string(content)

	const importLine = `"dlt-ingress/src/main/issuance/internal/app/query/getprocess"`
	if !strings.Contains(sContent, importLine) {
		idx := strings.Index(sContent, ")")
		if idx == -1 {
			fmt.Println("❌ Error: no import block found in", filePath)
			os.Exit(1)
		}
		sContent = sContent[:idx] + "\t" + importLine + "\n" + sContent[idx:]
	}

	const handlerLine = "getprocessquery.NewHandler(repositories.BaseProcessRepo),"
	if !strings.Contains(sContent, handlerLine) {
		marker := "return []any{"
		idx := strings.Index(sContent, marker)
		if idx == -1 {
			fmt.Printf("❌ Error: %q not found in %s\n", marker, filePath)
			os.Exit(1)
		}
		insertAt := idx + len(marker)
		sContent = sContent[:insertAt] + "\n\t\t" + handlerLine + sContent[insertAt:]
	}

	if err := os.WriteFile(filePath, []byte(sContent), 0644); err != nil {
		fmt.Println("❌ Error writing", filePath, ":", err)
		os.Exit(1)
	}
	fmt.Println("⚙️  Query handler wired into:", filePath)
}

// injectRouterRegistration wires GET /issuance/process/:processId into config/router.go's
// Query (GET) section.
func injectRouterRegistration(root string) {
	filePath := filepath.Join(root, "issuance/config/router.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ Error reading", filePath, ":", err)
		os.Exit(1)
	}
	sContent := string(content)

	const importLine = `"dlt-ingress/src/main/issuance/internal/infra/api/process/getprocess"`
	if !strings.Contains(sContent, importLine) {
		idx := strings.Index(sContent, ")")
		if idx == -1 {
			fmt.Println("❌ Error: no import block found in", filePath)
			os.Exit(1)
		}
		sContent = sContent[:idx] + "\t" + importLine + "\n" + sContent[idx:]
	}

	const registerLine = "registrar.Register(pathProtected.GET, getprocessapi.NewEndpoint(queryBus))"
	if !strings.Contains(sContent, registerLine) {
		marker := "// Query (GET)"
		idx := strings.Index(sContent, marker)
		if idx == -1 {
			fmt.Printf("❌ Error: marker %q not found in %s\n", marker, filePath)
			os.Exit(1)
		}
		lineEnd := strings.Index(sContent[idx:], "\n")
		if lineEnd == -1 {
			fmt.Println("❌ Error: could not find end of marker line in", filePath)
			os.Exit(1)
		}
		insertAt := idx + lineEnd + 1
		sContent = sContent[:insertAt] + "\t" + registerLine + "\n" + sContent[insertAt:]
	}

	if err := os.WriteFile(filePath, []byte(sContent), 0644); err != nil {
		fmt.Println("❌ Error writing", filePath, ":", err)
		os.Exit(1)
	}
	fmt.Println("⚙️  Route wired into:", filePath)
}
