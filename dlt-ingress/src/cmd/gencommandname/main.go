package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

var tmpl *template.Template
var completedOk = false

func init() {
	var err error
	tmpl, err = template.ParseFiles("./../cmd/gencommandname/method.tmpl")
	if err != nil {
		fmt.Println("❌ Failed to load template:", err)
		os.Exit(1)
	}
}

func starting() {
	fmt.Println("──────────────────────────────────────")
	fmt.Println("🏷️  Starting commandname generator")
	fmt.Println("──────────────────────────────────────")
}

func ending() {
	fmt.Println("───────────────────────────────")

	if completedOk {
		fmt.Println("✅ Commandname generation finished successfully")
	} else {
		fmt.Println("❌ Commandname generation failed")
	}
	fmt.Println("───────────────────────────────")
	fmt.Println()
	fmt.Println()
}

func main() {
	starting()
	defer ending()
	root := "./../main/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_gen.go") {
			return nil
		}
		return processFile(path)
	})
	if err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	} else {
		completedOk = true
	}
}

func processFile(path string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			// Forzamos que se llame Command
			if !ok || typeSpec.Name.Name != "Command" {
				continue
			}

			_, ok = typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			fullPath := getImportPath(filepath.Dir(path))

			return generateCode(path, node.Name.Name, fullPath)
		}
	}
	return nil
}

func getImportPath(dir string) string {
	out, err := exec.Command("go", "list", "-f", "{{.ImportPath}}", "./"+dir).Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func generateCode(origPath, packageName, fullPkgPath string) error {
	prefix := extractPrefix(origPath)
	var genName string
	if prefix != "" {
		genName = prefix + "_commandname_gen.go"
	} else {
		genName = "commandname_gen.go" //TODO: Replace by error when refactor completed in all BC
	}
	genPath := filepath.Join(filepath.Dir(origPath), genName)
	fmt.Println("⚙️  Generating:", genPath)

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, struct {
		Package     string
		FullPkgPath string
	}{
		Package:     packageName,
		FullPkgPath: fullPkgPath,
	})

	if err != nil {
		return err
	}

	return os.WriteFile(genPath, buf.Bytes(), 0644)
}

func extractPrefix(path string) string {
	const suffix = "_command.go"
	base := filepath.Base(path)
	if !strings.HasSuffix(base, suffix) {
		return ""
	}
	prefix := base[:len(base)-len(suffix)]
	if prefix == "" {
		return ""
	}
	return prefix
}
