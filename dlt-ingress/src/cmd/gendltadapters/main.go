package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

var completedOk = false
var anyError = false

type ParamDef struct {
	OriginalName string
	CleanName    string
	CleanType    string
}

type MethodDef struct {
	Name         string
	OriginalName string
	Params       []ParamDef
	IsTx         bool // true if TransactOpts
	ReturnType   string
	ZeroValue    string
}

type TemplateData struct {
	OriginalName  string // e.g., IERC1643
	InterfaceName string // e.g., IERC1643Agnostic
	AdapterName   string // e.g., ERC1643CrossAdapter
	MockName      string // e.g., IERC1643AgnosticMock
	Methods       []MethodDef
	NeedsBigInt   bool
}

func starting() {
	fmt.Println("======================================")
	fmt.Println("🚀 Starting DLT Adapters generator")
	fmt.Println("======================================")
}

func ending() {
	fmt.Println("===============================")
	if completedOk {
		fmt.Println("✅ DLT Adapters generation finished successfully")
	} else {
		fmt.Println("❌ DLT Adapters generation failed")
	}
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println()
}

func main() {
	starting()
	defer ending()

	sourceDir := "./core/blockchain/contracts"
	outDir := "./core/blockchain/contracts/agnostic"
	mockOutDir := "./../test/core/blockchain/contracts/agnostic"

	for _, dir := range []string{outDir, mockOutDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("❌ Error creating output directory: %v\n", err)
			anyError = true
			return
		}
	}

	tmpl, err := template.ParseFiles("./../cmd/gendltadapters/adapter.tmpl")
	if err != nil {
		fmt.Printf("❌ Error loading adapter template: %v\n", err)
		anyError = true
		return
	}

	mockTmpl, err := template.ParseFiles("./../cmd/gendltadapters/mock.tmpl")
	if err != nil {
		fmt.Printf("❌ Error loading mock template: %v\n", err)
		anyError = true
		return
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, sourceDir, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("❌ Error parsing directory: %v\n", err)
		anyError = true
		return
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					iface, ok := typeSpec.Type.(*ast.InterfaceType)
					if !ok {
						continue // Not an interface
					}

					originalName := typeSpec.Name.Name
					// Only process interfaces starting with 'I' (e.g., IERC1643)
					if !strings.HasPrefix(originalName, "I") {
						continue
					}

					processInterface(originalName, iface, outDir, mockOutDir, tmpl, mockTmpl, fset, file)
				}
			}
		}
	}

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error executing go fmt: %v", err)
	}

	if !anyError {
		completedOk = true
	}
}

func processInterface(originalName string, iface *ast.InterfaceType, outDir string, mockOutDir string, tmpl *template.Template, mockTmpl *template.Template, fset *token.FileSet, file *ast.File) {
	data := TemplateData{
		OriginalName:  originalName,
		InterfaceName: originalName + "Agnostic",
		AdapterName:   strings.TrimPrefix(originalName, "I") + "CrossAdapter",
		MockName:      originalName + "AgnosticMock",
	}

	for _, method := range iface.Methods.List {
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		methodDef := MethodDef{
			Name:         method.Names[0].Name,
			OriginalName: toCamelCase(method.Names[0].Name),
		}

		for _, param := range funcType.Params.List {
			paramTypeStr := getTypeString(param.Type, file)

			// Detect if it's a write transaction and skip the opts parameter
			if strings.Contains(paramTypeStr, "bind.TransactOpts") {
				methodDef.IsTx = true
				continue
			}
			if strings.Contains(paramTypeStr, "bind.CallOpts") {
				continue
			}

			var originalParamName string
			if len(param.Names) > 0 {
				originalParamName = param.Names[0].Name
			} else {
				originalParamName = "arg"
			}

			cleanParamName := strings.TrimPrefix(originalParamName, "_")
			if cleanParamName == "type" {
				cleanParamName = "typ" // Avoid Go reserved word
			}

			// Type mapping
			cleanType := paramTypeStr
			// Example: common.Address to be string:
			if cleanType == "common.Address" {
				cleanType = "string"
			}

			methodDef.Params = append(methodDef.Params, ParamDef{
				OriginalName: originalParamName,
				CleanName:    cleanParamName,
				CleanType:    cleanType,
			})
		}

		if !methodDef.IsTx && funcType.Results != nil && len(funcType.Results.List) > 0 {
			retTypeStr := getTypeString(funcType.Results.List[0].Type, file)
			if retTypeStr == "common.Address" {
				retTypeStr = "string"
			}
			methodDef.ReturnType = retTypeStr
			methodDef.ZeroValue = zeroValueFor(retTypeStr)
			if strings.Contains(retTypeStr, "big.Int") {
				data.NeedsBigInt = true
			}
		}

		data.Methods = append(data.Methods, methodDef)
	}

	if len(data.Methods) == 0 {
		return
	}

	fileName := strings.ToLower(strings.TrimPrefix(originalName, "I")) + "_adapter_gen.go"
	outPath := filepath.Join(outDir, fileName)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		fmt.Printf("❌ Error generating %s: %v\n", originalName, err)
		anyError = true
		return
	}

	if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
		fmt.Printf("❌ Error writing file %s: %v\n", outPath, err)
		anyError = true
		return
	}

	fmt.Printf("📝 Generated: %s\n", outPath)

	mockFileName := strings.ToLower(strings.TrimPrefix(originalName, "I")) + "_agnostic_mock_gen.go"
	mockOutPath := filepath.Join(mockOutDir, mockFileName)

	var mockBuf bytes.Buffer
	if err := mockTmpl.Execute(&mockBuf, data); err != nil {
		fmt.Printf("❌ Error generating mock %s: %v\n", originalName, err)
		anyError = true
		return
	}

	if err := os.WriteFile(mockOutPath, mockBuf.Bytes(), 0644); err != nil {
		fmt.Printf("❌ Error writing file %s: %v\n", mockOutPath, err)
		anyError = true
		return
	}

	fmt.Printf("📝 Generated: %s\n", mockOutPath)
}

// Helper function to extract AST type to a readable string
func getTypeString(expr ast.Expr, file *ast.File) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return getTypeString(t.X, file) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X, file)
	case *ast.ArrayType:
		lenStr := ""
		if t.Len != nil {
			if basicLit, ok := t.Len.(*ast.BasicLit); ok {
				lenStr = basicLit.Value
			}
		}
		return "[" + lenStr + "]" + getTypeString(t.Elt, file)
	case *ast.SliceExpr:
		return "[]" + getTypeString(t.X, file)
	default:
		return "any"
	}
}

func zeroValueFor(t string) string {
	if strings.HasPrefix(t, "*") || strings.HasPrefix(t, "[]") {
		return "nil"
	}
	switch t {
	case "string":
		return `""`
	case "bool":
		return "false"
	default:
		return "0"
	}
}

// Helper function to convert to lowerCamelCase (e.g., SetDocument -> setDocument)
func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
