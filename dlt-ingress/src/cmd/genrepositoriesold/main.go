package main

import (
	"bytes"
	cmdcore "dlt-ingress/src/cmd/core"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var tmpl, tmplTest *template.Template
var completedOk = false

func init() {
	var err error
	titleCaser := cases.Title(language.Und)

	funcs := template.FuncMap{
		"EntityName":    entityName,
		"EntityPackage": entityPackage,
		"Title":         titleCaser.String,
	}
	tmpl, err = template.New("method.tmpl").Funcs(funcs).ParseFiles("./../cmd/genrepositoriesold/method.tmpl")
	if err != nil {
		fmt.Println("❌ Failed to load template:", err)
		os.Exit(1)
	}
	tmplTest, err = template.New("mock.tmpl").Funcs(funcs).ParseFiles("./../cmd/genrepositoriesold/mock.tmpl")
	if err != nil {
		fmt.Println("❌ Failed to load template:", err)
		os.Exit(1)
	}

}

func starting() {
	fmt.Println("──────────────────────────────────────")
	fmt.Println("🛢️ Starting repositories generator")
	fmt.Println("──────────────────────────────────────")
}

func ending() {
	fmt.Println("───────────────────────────────")

	if completedOk {
		fmt.Println("✅ Repositories generation finished successfully")
	} else {
		fmt.Println("❌ Repositories generation failed")
	}
	fmt.Println("───────────────────────────────")
	fmt.Println()
	fmt.Println()
}

func main() {
	starting()
	defer ending()

	rootBase := "./../main/"
	anyError := false

	for _, bc := range cmdcore.BOUNDED_CONTEXTS {
		if cmdcore.IsDisabled(bc) {
			fmt.Printf("-- Skipping %s: Disabled\n", bc)
			continue
		}
		if cmdcore.IsMigrated(bc) {
			fmt.Printf("-- Skipping %s: Handled by new generator\n", bc)
			continue
		}

		fmt.Printf(">> Processing Bounded Context: %s\n", bc)

		bcPath := filepath.Join(rootBase, bc)

		if _, err := os.Stat(bcPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(bcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			return processPackage(bc, path)
		})

		if err != nil {
			fmt.Printf("? Error processing BC %s: %v\n", bc, err)
			anyError = true
		}
	}

	if !anyError {
		completedOk = true
	}
}

func processPackage(bc, dir string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_gen.go") && strings.HasSuffix(fi.Name(), ".go")
	}, parser.ParseComments)
	if err != nil {
		return err
	}

	for pkgName, pkg := range pkgs {
		// Search for repository interfaces
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}

				for _, spec := range gen.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || !strings.HasSuffix(typeSpec.Name.Name, "Repository") || strings.HasPrefix(typeSpec.Name.Name, "Base") {
						continue
					}

					interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
					if !ok {
						continue
					}

					secondType, ok := extractSecondGenericType(interfaceType)
					if !ok {
						continue
					}

					domainImport, ok := extractDomainImport(file, secondType)
					if !ok {
						continue
					}

					// Search implementation struct
					for _, structFile := range pkg.Files {
						fileName := fset.File(structFile.Pos()).Name()
						if !strings.Contains(fileName, "postgres_") {
							continue
						}
						for _, structDecl := range structFile.Decls {
							structGen, ok := structDecl.(*ast.GenDecl)
							if !ok || structGen.Tok != token.TYPE {
								continue
							}

							for _, spec := range structGen.Specs {
								typeSpec, ok := spec.(*ast.TypeSpec)
								if !ok || !strings.HasSuffix(typeSpec.Name.Name, "Repository") {
									continue
								}

								if _, ok := typeSpec.Type.(*ast.StructType); !ok {
									continue
								}

								// Generate code
								err := generateCode(bc, fileName, pkgName, typeSpec.Name.Name, domainImport, secondType)
								if err != nil {
									return err
								}
								return generateMock(bc, fileName, pkgName, typeSpec.Name.Name, domainImport, secondType)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func extractDomainImport(file *ast.File, secondType string) (string, bool) {
	packageImportName := strings.Split(secondType, ".")[0]
	var domainImport string
	const domainImportCommonPath = "domain/"
	for _, imp := range file.Imports {
		if strings.Contains(imp.Path.Value, domainImportCommonPath) {
			suffixImport := strings.SplitN(strings.Replace(imp.Path.Value, "\"", "", -1), domainImportCommonPath, 2)[1]
			suffixImportParts := strings.Split(suffixImport, "/")
			var combined string
			for i := len(suffixImportParts) - 1; i >= 0; i-- {
				if combined == "" {
					combined = suffixImportParts[i]
				} else {
					combined = combined + suffixImportParts[i]
				}
				if combined == packageImportName {
					domainImport = imp.Path.Value
					break
				}
			}
		}
	}
	if domainImport == "" {
		return "", false
	}
	return domainImport, true
}

func extractSecondGenericType(iface *ast.InterfaceType) (string, bool) {
	for _, method := range iface.Methods.List {
		if method.Names != nil {
			continue
		}

		switch expr := method.Type.(type) {
		case *ast.IndexListExpr:
			if len(expr.Indices) == 2 {
				switch sel := expr.Indices[1].(type) {
				case *ast.SelectorExpr:
					if pkgIdent, ok := sel.X.(*ast.Ident); ok {
						return pkgIdent.Name + "." + sel.Sel.Name, true
					}
				case *ast.Ident:
					return sel.Name, true
				}
			}
		}
	}
	return "", false
}

func generateCode(bc, origPath, packageName, typeName, domainImport, entityNamePackage string) error {
	prefix := extractPrefix(origPath)
	genName := prefix + "commonrepo_gen.go"
	domainErrorImport := getDomainErrorsPath(bc, domainImport)
	genPath := filepath.Join(filepath.Dir(origPath), genName)
	fmt.Println("⚙️  Generating:", genPath)

	imports := []string{
		"context",
		"fmt",
		"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db",
		"errors",
		"github.com/google/uuid",
		"gorm.io/gorm",
		"gorm.io/gorm/clause",
		"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination",
		strings.Replace(domainImport, "\"", "", -1),
	}

	if domainErrorImport != "" {
		imports = append(imports, strings.Replace(domainErrorImport, "\"", "", -1))
	}

	sort.Strings(imports)
	entityName := strings.Split(entityNamePackage, ".")[1]

	var notFoundDomainError string
	if cmdcore.IsErrorRefactor(bc) {
		notFoundDomainError = fmt.Sprintf("domainerrors.NewEntityNotFoundDomainError(\"%s\", id)", entityName)
	} else {
		notFoundDomainError = fmt.Sprintf("domainerrors.NewEntityNotFoundError(\"%s\", id, nil)", entityName)
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, struct {
		Package             string
		Repository          string
		DomainImport        string
		Entity              string
		EntityName          string
		Imports             []string
		NotFoundDomainError string
	}{
		Package:             packageName,
		Repository:          typeName,
		DomainImport:        domainImport,
		Entity:              entityNamePackage,
		EntityName:          entityName,
		Imports:             imports,
		NotFoundDomainError: notFoundDomainError,
	})

	if err != nil {
		return err
	}
	return os.WriteFile(genPath, buf.Bytes(), 0644)
}

func generateMock(bc, origPath, packageName, typeName, domainImport, entityNamePackage string) error {
	prefix := extractPrefix(origPath)
	domain := extractDomain(origPath)
	genName := prefix + "repository_mock_gen.go"
	genPath := filepath.Join("../test", domain, "port/repository", genName)
	entityRepoImport := "\"dlt-ingress/src/" + extractPath(origPath) + "\""
	domainImport = strings.Replace(domainImport, "\"", "", -1)

	fmt.Println("⚙️  Generating:", genPath)
	imports := []string{
		"\"context\"",
		"\"github.com/google/uuid\"",
		"\"github.com/stretchr/testify/mock\"",
		"\"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination\"",
		entityRepoImport,
		"\"" + domainImport + "\"",
	}

	sort.Strings(imports)
	entityName := strings.Split(entityNamePackage, ".")[1]
	var buf bytes.Buffer
	err := tmplTest.Execute(&buf, struct {
		Package      string
		Repository   string
		DomainImport string
		Entity       string
		EntityName   string
		Imports      []string
	}{
		Package:      packageName,
		Repository:   typeName,
		DomainImport: domainImport,
		Entity:       entityNamePackage,
		EntityName:   entityName,
		Imports:      imports,
	})

	if err != nil {
		return err
	}
	return os.WriteFile(genPath, buf.Bytes(), 0644)
}

func getDomainErrorsPath(bc, importPath string) string {
	if cmdcore.IsErrorRefactor(bc) {
		return "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	} else {
		path := strings.Trim(importPath, "\" ")

		if strings.Contains(path, "main/core/domain/baseprocess") {
			return fmt.Sprintf("dlt-ingress/src/main/%s/domain/domainerrors", bc)
		}

		re := regexp.MustCompile(`(dlt-ingress/src/main/)([^/]+)/((internal/)?domain)`)
		matches := re.FindStringSubmatch(path)

		if len(matches) < 4 {
			return ""
		}

		prefix := matches[1]
		boundedContext := matches[2]
		domainPart := matches[3]

		return fmt.Sprintf("%s%s/%s/domainerrors", prefix, boundedContext, domainPart)
	}
}

func extractDomain(path string) string {
	trimmed := strings.TrimPrefix(path, "../main/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func entityName(entity string) string {
	parts := strings.SplitN(entity, ".", 2)
	if len(parts) > 0 {
		return strings.ToLower(parts[1])
	}
	return ""
}

func entityPackage(entity string) string {
	if entity == "entity.Account" { //hack for special entity placement
		return "account"
	} else if entity == "baseprocess.BaseProcess" {
		return "process"
	}
	parts := strings.SplitN(entity, ".", 2)
	if len(parts) > 0 {
		return strings.ToLower(parts[0])
	}
	return ""
}

func extractPrefix(path string) string {
	const suffix = "postgres_repository.go"
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

func extractPath(path string) string {
	trimmed := strings.TrimPrefix(path, "../")
	idx := strings.LastIndex(trimmed, "/")
	if idx != -1 {
		return trimmed[:idx]
	}
	return trimmed
}
